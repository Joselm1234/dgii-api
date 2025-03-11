package contributors

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	// Asegúrate de tener este paquete o crea uno similar
	"my-dgii-api/bunapp"
	"my-dgii-api/httputil"

	"github.com/google/uuid"
	"github.com/ulikunitz/xz"
	"github.com/uptrace/bunrouter"
)

type ContributorHandler struct {
	app *bunapp.App
}

// NewContributorHandler crea un nuevo ContributorHandler.
func NewContributorHandler(app *bunapp.App) *ContributorHandler {
	return &ContributorHandler{
		app: app,
	}
}

// GetContributors maneja la solicitud para obtener todos los contribuyentes.
func (h *ContributorHandler) GetContributors(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	var contributors []*Contributor
	count, err := h.app.DB().NewSelect().
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

	contributor, err := SelectContributorByRNC(ctx, h.app, rnc)
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
	err := h.app.DB().NewSelect().Model(contributor).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return err
	}

	return bunrouter.JSON(w, contributor)
}

// CreateContributor maneja la solicitud para crear un nuevo contribuyente.
func (h *ContributorHandler) CreateContributor(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	var contributor Contributor
	
	if err := httputil.BindJSON(w, req, &contributor); err != nil {
		return err
	}
    tx, err := h.app.DB().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	if err := contributor.Save(ctx, tx); err != nil{
		tx.Rollback()
		return err
	}

	return bunrouter.JSON(w, contributor)
}

// // UpdateContributor maneja la solicitud para actualizar un contribuyente existente.
// func (h *ContributorHandler) UpdateContributor(w http.ResponseWriter, req bunrouter.Request) error {
// 	ctx := req.Context()
// 	rnc := req.Param("rnc")
// 	contributor, err := SelectContributorByRNC(ctx, h.db, rnc)
// 	if err != nil {
// 		return err
// 	}
    
// 	tx, err := h.db.BeginTx(ctx, &sql.TxOptions{})
// 	if err != nil {
// 		return err
// 	}
// 	if err := contributor.Save(ctx, tx); err != nil {
// 		tx.Rollback()
// 		return err
// 	}

// 	err = contributor.Save(ctx, h.db)
// 	if err != nil {
// 		return err
// 	}

// 	return bunrouter.JSON(w, contributor)
// }

// // DeleteContributor maneja la solicitud para eliminar un contribuyente por ID.
// func (h *ContributorHandler) DeleteContributor(w http.ResponseWriter, req bunrouter.Request) error {
// 	ctx := req.Context()
// 	id := req.Param("id")
// 	contributor := new(Contributor)
// 	_, err := h.db.NewDelete().Model(contributor).Where("id = ?", id).Exec(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	return bunrouter.JSON(w, bunrouter.H{
// 		"success": "true",
// 	})
// }
// ImportContributors maneja la solicitud para importar contribuyentes desde un archivo ZIP de la DGII.
func (h *ContributorHandler) ImportContributors(w http.ResponseWriter, req bunrouter.Request) error {
    if err := h.ImportContributorsFromDGII(); err != nil {
        return err
    }

    return bunrouter.JSON(w, bunrouter.H{"success": "true"})
}

func (h *ContributorHandler) ImportContributorsFromDGII() error {
    ctx := context.Background()

    reader, err := h.getContributorFileReader()
    if err != nil {
        return err
    }
    defer reader.Close()

    scanner := bufio.NewScanner(reader)
    batchSize := 1000
    contributors := make([]*Contributor, 0, batchSize)

    for scanner.Scan() {
        line := scanner.Text()
        contributor, err := h.parseContributorFromLine(line)
        if err != nil {
            log.Printf("error al procesar la línea: %v", err)
            continue
        }
        contributors = append(contributors, contributor)

        if len(contributors) >= batchSize {
            if err := h.insertContributorsBatch(ctx, contributors); err != nil {
                return err
            }
            contributors = contributors[:0]
        }
    }

    if len(contributors) > 0 {
        if err := h.insertContributorsBatch(ctx, contributors); err != nil {
            return err
        }
    }

    if err := scanner.Err(); err != nil {
        return fmt.Errorf("error al leer el archivo: %w", err)
    }

    return nil
}

func (h *ContributorHandler) getContributorFileReader() (io.ReadCloser, error) {
    resp, err := http.Get("https://dgii.gov.do/app/WebApps/Consultas/RNC/DGII_RNC.zip")
    if err != nil {
        return nil, fmt.Errorf("error al descargar el ZIP: %w", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("error al leer el cuerpo de la respuesta: %w", err)
    }

    zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
    if err != nil {
        return nil, fmt.Errorf("error al leer el ZIP: %w", err)
    }

    var txtFile *zip.File // Declaración de txtFile
    for _, file := range zipReader.File {
        if file.Name == "TMP/DGII_RNC.TXT" {
            txtFile = file // Asignación de txtFile
            break
        }
    }

    if txtFile == nil { // Uso de txtFile
        return nil, fmt.Errorf("no se encontró el archivo DGII_RNC.TXT")
    }

    rc, err := txtFile.Open() // Uso de txtFile
    if err != nil {
        return nil, fmt.Errorf("error al abrir el archivo TXT: %w", err)
    }

    var reader io.Reader = rc
    if strings.HasSuffix(txtFile.Name, ".xz") { // Uso de txtFile
        xzReader, err := xz.NewReader(rc)
        if err != nil {
            return nil, fmt.Errorf("error al descomprimir el archivo: %w", err)
        }
        reader = xzReader
    }

    return reader.(io.ReadCloser), nil
}

func (h *ContributorHandler) insertContributorsBatch(ctx context.Context, contributors []*Contributor) error {
    tx, err := h.app.DB().BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    for _, c := range contributors {
        c.ID = uuid.NewString()
        _, err := tx.NewInsert().Model(c).Exec(ctx)
        if err != nil {
            return err
        }
    }

    return tx.Commit()
}

func (h *ContributorHandler) parseContributorFromLine(line string) (*Contributor, error) {
    fields := strings.Split(line, "|")
    if len(fields) < 10 {
        return nil, fmt.Errorf("línea inválida: %s", line)
    }

    startDateStr := strings.TrimSpace(fields[8])
    var startDate time.Time
    var err error

    if startDateStr != "" {
        startDate, err = time.Parse("02/01/2006", startDateStr)
        if err != nil {
            return nil, fmt.Errorf("error al parsear la fecha: %w", err)
        }
    } else {
        startDate = time.Time{}
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
// func (h *ContributorHandler) insertOrUpdateContributor(ctx context.Context, db *bun.DB, contributor *Contributor) error {
// 	// 1. Verificar si el contribuyente ya existe por RNC
// 	// existingContributor, err := SelectContributorByRNC(ctx, db, contributor.RNC)
// 	// if err != nil && err != sql.ErrNoRows {
// 	// 	return fmt.Errorf("error al verificar el contribuyente: %w", err)
// 	// }

// 	// 2. Insertar o actualizar
// 	// if existingContributor == nil {
// 		// Insertar si no existe
// 		err := contributor.Save(ctx, db)
// 		if err != nil {
// 			return fmt.Errorf("error al insertar el contribuyente: %w", err)
// 		}
// 	// }

// 	return nil
// }
