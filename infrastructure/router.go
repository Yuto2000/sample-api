package infrastructure

import (
	"context"
	"net/http"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	handlers "github.com/uma-arai/sbcntr-backend/handler"
	"github.com/uma-arai/sbcntr-backend/utils"
)

// Router ...
func Router() *echo.Echo {
	e := echo.New()
	rd := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
	})

	// Middleware
	logger := middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"id":"${id}","time":"${time_rfc3339}","remote_ip":"${remote_ip}",` +
			`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}",` +
			`"status":${status},"error":"${error}"}` + "\n",
		Output: os.Stdout,
	})
	e.Use(logger)
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.INFO)
	e.HideBanner = true
	e.HidePort = false

	AppHandler := handlers.NewAppHandler(NewSQLHandler())
	healthCheckHandler := handlers.NewHealthCheckHandler()
	helloWorldHandler := handlers.NewHelloWorldHandler()
	NotificationHandler := handlers.NewNotificationHandler(NewSQLHandler())

	e.GET("/", healthCheckHandler.HealthCheck())
	e.GET("/healthcheck", healthCheckHandler.HealthCheck())
	e.GET("/v1/helloworld", helloWorldHandler.SayHelloWorld())

	e.GET("/v1/Items", AppHandler.GetItems())
	e.POST("/v1/Item", AppHandler.CreateItem())
	e.POST("/v1/Item/favorite", AppHandler.UpdateFavoriteAttr())

	e.GET("/v1/Notifications", NotificationHandler.GetNotifications())
	e.GET("/v1/Notifications/Count", NotificationHandler.GetUnreadNotificationCount())
	e.POST("/v1/Notifications/Read", NotificationHandler.PostNotificationsRead())

	e.GET("/set", set(*rd))
	e.GET("/get", get(*rd))

	return e
}

func set(rd redis.Client) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		sessionID := c.QueryParam("session")
		value := c.QueryParam("value")
		if err := rd.Set(context.TODO(), sessionID, value, 0).Err(); err != nil {
			return utils.GetErrorMassage(c, "en", err)
		}
		resJSON := Message{
			Message: "success",
		}
		return c.JSON(http.StatusOK, resJSON)
	}
}

func get(rd redis.Client) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		sessionID := c.QueryParam("session")
		value, err := rd.Get(context.TODO(), sessionID).Result()
		if err != nil {
			return utils.GetErrorMassage(c, "en", err)
		}
		resJSON := SessionValue{
			SesionValue: value,
		}
		return c.JSON(http.StatusOK, resJSON)
	}
}

type (
	Message struct {
		Message string `json:"message"`
	}

	SessionValue struct {
		SesionValue string `json:"sessionValue"`
	}
)
