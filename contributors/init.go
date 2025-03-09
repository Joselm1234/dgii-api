package contributors

import (
	"my-dgii-api/database"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
)

func InitRoutes(db *bun.DB) *bunrouter.Router {
    router := bunrouter.New()
    handler := NewContributorHandler(db)

    // Rutas para Contribuyentes
    router.GET("/contributors", handler.GetContributors)
    router.GET("/contributors/:rnc", handler.GetByRnc)
    router.POST("/contributors", handler.CreateContributor)
    router.PUT("/contributors/:id", handler.UpdateContributor)
    router.DELETE("/contributors/:id", handler.DeleteContributor)
    router.POST("/contributors/import", handler.ImportContributors)

    return router
}

func InitModule() *bunrouter.Router {
    db := database.ConnectDB()
	
    router := InitRoutes(db)
    return router
}