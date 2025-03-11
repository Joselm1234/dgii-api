package contributors

import (
	"context"

	"my-dgii-api/bunapp"
)

func init() {
	bunapp.OnStart("contributor.initRoutes", func(ctx context.Context, app *bunapp.App) error {
		app.DB().RegisterModel((*Contributor)(nil))

		contributorHandler := NewContributorHandler(app)

		g := app.APIRouter().NewGroup("/v1")


        g.GET("/contributors", contributorHandler.GetContributors)
        g.GET("/contributors/:rnc", contributorHandler.GetByRnc)
        g.POST("/contributors", contributorHandler.CreateContributor)

		g.POST("/contributors/import", contributorHandler.ImportContributors)

		return nil
	})
}
