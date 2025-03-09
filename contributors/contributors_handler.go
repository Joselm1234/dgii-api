package contributors

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	// Asegúrate de tener este paquete o crea uno similar
	"github.com/ulikunitz/xz/lzma"
	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
)

type ContributorHandler struct {
	db *bun.DB
}

// NewContributorHandler crea un nuevo ContributorHandler.
func NewContributorHandler(db *bun.DB) *ContributorHandler {
	return &ContributorHandler{
		db: db,
	}
}

// GetContributors maneja la solicitud para obtener todos los contribuyentes.
func (h *ContributorHandler) GetContributors(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	var contributors []*Contributor
	count, err := h.db.NewSelect().
		Model(&contributors).
		Order("created_at DESC").
		ScanAndCount(ctx)
	if err != nil {
		return err
	}

	return bunrouter.JSON(w, bunrouter.H{
		"rows":       contributors,
		"totalCount": count,
	})
}

// GetByRnc maneja la solicitud para obtener un contribuyente por RNC.
func (h *ContributorHandler) GetByRnc(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	rnc := req.Param("rnc")

	contributor, err := SelectContributorByRNC(ctx, h.db, rnc)
	if err != nil {
		return err
	}

	return bunrouter.JSON(w, contributor)
}

// GetContributor maneja la solicitud para obtener un contribuyente por ID.
func (h *ContributorHandler) GetContributor(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	id := req.Param("id")
	contributor := new(Contributor)
	err := h.db.NewSelect().Model(contributor).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return err
	}

	return bunrouter.JSON(w, contributor)
}

// CreateContributor maneja la solicitud para crear un nuevo contribuyente.
func (h *ContributorHandler) CreateContributor(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	contributor := new(Contributor)
	if err := json.NewDecoder(req.Body).Decode(&contributor); err != nil {
		return err
	}

	err := contributor.Save(ctx, h.db)
	if err != nil {
		return err
	}

	return bunrouter.JSON(w, contributor)
}

// UpdateContributor maneja la solicitud para actualizar un contribuyente existente.
func (h *ContributorHandler) UpdateContributor(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	rnc := req.Param("rnc")
	contributor, err := SelectContributorByRNC(ctx, h.db, rnc)
	if err != nil {
		return err
	}
    
	tx, err := h.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	if err := contributor.Save(ctx, tx); err != nil {
		tx.Rollback()
		return err
	}

	err = contributor.Save(ctx, h.db)
	if err != nil {
		return err
	}

	return bunrouter.JSON(w, contributor)
}

// DeleteContributor maneja la solicitud para eliminar un contribuyente por ID.
func (h *ContributorHandler) DeleteContributor(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	id := req.Param("id")
	contributor := new(Contributor)
	_, err := h.db.NewDelete().Model(contributor).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return err
	}

	return bunrouter.JSON(w, bunrouter.H{
		"success": "true",
	})
}
// ImportContributors maneja la solicitud para importar contribuyentes desde un archivo ZIP de la DGII.
func (h *ContributorHandler) ImportContributors(w http.ResponseWriter, req bunrouter.Request) error {
	// ctx := req.Context()

	if err := h.ImportContributorsFromDGII(h.db); err != nil {
		return err
	}

	return bunrouter.JSON(w, bunrouter.H{
		"success": "true",
	})
}
func (h *ContributorHandler) ImportContributorsFromDGII(db *bun.DB) error {
    ctx := context.Background()

    // 1. Descargar el ZIP
    resp, err := http.Get("https://dgii.gov.do/app/WebApps/Consultas/RNC/DGII_RNC.zip")
    if err != nil {
        return fmt.Errorf("error al descargar el ZIP: %w", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body) // Use io.ReadAll instead of ioutil.ReadAll
    if err != nil {
        return fmt.Errorf("error al leer el cuerpo de la respuesta: %w", err)
    }

    // 2. Leer el ZIP desde la memoria
    zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
    if err != nil {
        return fmt.Errorf("error al leer el ZIP: %w", err)
    }

    // 3. Buscar el archivo DGII_RNC.TXT dentro de la carpeta DGII_RNC/TMP
    var txtFile *zip.File
for _, file := range zipReader.File {
    fmt.Println("Archivo en el ZIP:", file.Name)
    if file.Name == "TMP/DGII_RNC.TXT" {
        txtFile = file
        break
    }
}


    if txtFile == nil {
        return fmt.Errorf("no se encontró el archivo DGII_RNC.TXT dentro de la carpeta DGII_RNC/TMP/")
    }

    // 4. Abrir el archivo DGII_RNC.TXT dentro de la carpeta TMP
    rc, err := txtFile.Open()
    if err != nil {
        return fmt.Errorf("error al abrir el archivo %s: %w", txtFile.Name, err)
    }
    defer rc.Close()

    // 5. Descomprimir el archivo .xz (si es necesario)
	var reader io.Reader = rc // Inicializar con el ReadCloser del archivo ZIP
    if strings.HasSuffix(txtFile.Name, ".xz") {
        xzReader, err := lzma.NewReader(rc)
        if err != nil {
            return fmt.Errorf("error al descomprimir el archivo %s: %w", txtFile.Name, err)
        }
        reader = xzReader // Usar el lector descomprimido si es un archivo .xz
    }

    // 6. Leer el contenido del archivo línea por línea
    scanner := bufio.NewScanner(reader)
    for scanner.Scan() {
        line := scanner.Text()
        // 7. Procesar la línea y crear un Contributor
        contributor, err := h.parseContributorFromLine(line)
        if err != nil {
            log.Printf("error al procesar la línea: %v", err)
            continue // Continuar con la siguiente línea
        }

        // 8. Insertar o actualizar el Contributor en la base de datos
        err = h.insertOrUpdateContributor(ctx, db, contributor)
        if err != nil {
            log.Printf("error al insertar/actualizar el contribuyente: %v", err)
        }
    }

    if err := scanner.Err(); err != nil {
        return fmt.Errorf("error al leer el archivo %s: %w", txtFile.Name, err)
    }

    return nil
}
func (h *ContributorHandler) parseContributorFromLine(line string) (*Contributor, error) {
	log.Printf("Parsing line: %s", line) // Agregar este log
    fields := strings.Split(line, "|")
    if len(fields) < 10 {
        return nil, fmt.Errorf("línea inválida: %s", line)
    }

	startDateStr := strings.TrimSpace(fields[8]) // Eliminar espacios en blanco
    var startDate time.Time
    var err error

    if startDateStr != "" {
        startDate, err = time.Parse("02/01/2006", startDateStr)
        if err != nil {
            return nil, fmt.Errorf("error al parsear la fecha: %w", err)
        }
    } else {
        // Manejar fecha vacía (puedes asignar un valor predeterminado o omitir la línea)
        log.Printf("Fecha vacía en la línea: %s", line)
        startDate = time.Time{} // Asignar fecha nula
    }


    return &Contributor{
        RNC:                   fields[0],
        BusinessName:          fields[1],
		CommercialName:        fields[2],
        EconomicActivity:      fields[3],
		StartDateOfOperations: startDate,
        State:                 fields[9],

    }, nil
}

func (h *ContributorHandler) insertOrUpdateContributor(ctx context.Context, db *bun.DB, contributor *Contributor) error {
	// 1. Verificar si el contribuyente ya existe por RNC
	// existingContributor, err := SelectContributorByRNC(ctx, db, contributor.RNC)
	// if err != nil && err != sql.ErrNoRows {
	// 	return fmt.Errorf("error al verificar el contribuyente: %w", err)
	// }

	// 2. Insertar o actualizar
	// if existingContributor == nil {
		// Insertar si no existe
		err := contributor.Save(ctx, db)
		if err != nil {
			return fmt.Errorf("error al insertar el contribuyente: %w", err)
		}
	// }

	return nil
}
