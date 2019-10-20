// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"net/http"

	"github.com/go-openapi/swag"
	"github.com/jinzhu/gorm"

	errors "github.com/go-openapi/errors"
	runtime "github.com/go-openapi/runtime"
	middleware "github.com/go-openapi/runtime/middleware"

	"todo-list/dbmodels"
	"todo-list/models"
	"todo-list/restapi/operations"
	"todo-list/restapi/operations/todos"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

//go:generate swagger generate server --target ../../todo-list --name TodoList --spec ../swagger.yml

func configureFlags(api *operations.TodoListAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func addItem(item *models.Item) error {
	if item == nil {
		return errors.New(500, "item must be present")
	}

	var dbItem dbmodels.Item
	if item.Description != nil {
		dbItem.Description = dbmodels.NullString(*item.Description)
	}
	dbItem.Completed = item.Completed
	if err := db.Create(&dbItem).Error; err != nil {
		return errors.New(500, "cannot create new item: "+err.Error())
	}

	return nil
}

func updateItem(id int64, item *models.Item) error {
	if item == nil {
		return errors.New(500, "item must be present")
	}

	var dbItem dbmodels.Item
	if err := db.Find(&dbItem, id).Error; err != nil {
		return errors.NotFound("not found: item %d", id)
	}

	if item.Description != nil {
		dbItem.Description = dbmodels.NullString(*item.Description)
	}
	dbItem.Completed = item.Completed
	if err := db.Save(&dbItem).Error; err != nil {
		return errors.New(500, "cannot update item %d: "+err.Error(), id)
	}

	return nil
}

func deleteItem(id int64) error {
	var dbItem dbmodels.Item
	if err := db.Find(&dbItem, id).Error; err != nil {
		return errors.NotFound("not found: item %d", id)
	}

	if err := db.Delete(&dbItem).Error; err != nil {
		return errors.New(500, "cannot delete item %d: "+err.Error(), id)
	}

	return nil
}

func allItems(since int64, limit int32) (result []*models.Item) {
	result = make([]*models.Item, 0)
	var dbItems []dbmodels.Item
	if err := db.Limit(limit).Offset(since).Find(&dbItems).Error; err != nil {
		return
	}

	for _, v := range dbItems {
		var item models.Item
		item.ID = v.ID
		if v.Description.Valid {
			s := v.Description.String
			item.Description = &s
		}
		item.Completed = v.Completed
		result = append(result, &item)
	}
	return
}

var db *gorm.DB

func configureAPI(api *operations.TodoListAPI) http.Handler {
	var err error
	db, err = gorm.Open("postgres", "postgres://weerasak@localhost/todo_list?sslmode=disable")
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&dbmodels.Item{})
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	api.TodosAddOneHandler = todos.AddOneHandlerFunc(func(params todos.AddOneParams) middleware.Responder {
		if err := addItem(params.Body); err != nil {
			return todos.NewAddOneDefault(500).WithPayload(&models.Error{Code: 500, Message: swag.String(err.Error())})
		}
		return todos.NewAddOneCreated().WithPayload(params.Body)
	})

	api.TodosDestroyOneHandler = todos.DestroyOneHandlerFunc(func(params todos.DestroyOneParams) middleware.Responder {
		if err := deleteItem(params.ID); err != nil {
			return todos.NewDestroyOneDefault(500).WithPayload(&models.Error{Code: 500, Message: swag.String(err.Error())})
		}
		return todos.NewDestroyOneNoContent()
	})

	api.TodosFindTodosHandler = todos.FindTodosHandlerFunc(func(params todos.FindTodosParams) middleware.Responder {
		mergedParams := todos.NewFindTodosParams()
		mergedParams.Since = swag.Int64(0)
		if params.Since != nil {
			mergedParams.Since = params.Since
		}
		if params.Limit != nil {
			mergedParams.Limit = params.Limit
		}
		return todos.NewFindTodosOK().WithPayload(allItems(*mergedParams.Since, *mergedParams.Limit))
	})

	api.TodosUpdateOneHandler = todos.UpdateOneHandlerFunc(func(params todos.UpdateOneParams) middleware.Responder {
		if err := updateItem(params.ID, params.Body); err != nil {
			return todos.NewUpdateOneDefault(500).WithPayload(&models.Error{Code: 500, Message: swag.String(err.Error())})
		}
		return todos.NewUpdateOneOK().WithPayload(params.Body)
	})

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix"
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
