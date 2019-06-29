package main

import (
	"fmt"

	"github.com/kukinsula/boxy/entity"
	"github.com/kukinsula/boxy/framework/api/client"
)

func main() {
	logger := entity.CleanMetaLogger(entity.StdoutLogger)
	service := client.NewService("http://127.0.0.1:8000",
		func(req *client.Request) {
			logger(entity.Log{
				UUID:    req.UUID,
				Level:   "debug",
				Message: fmt.Sprintf("HTTP %s %s", req.Method, req.URL),
				Meta: map[string]interface{}{
					"headers": req.Headers,
					"body":    req.Body,
				},
			})
		},

		func(resp *client.Response) {
			// meta := map[string]interface{}{"duration": resp.Duration}

			// if resp.Error == nil {
			// 	meta["headers"] = resp.Headers
			// 	meta["status"] = resp.Status
			// } else {
			// 	meta["error"] = resp.Error.Error()
			// }

			logger(entity.Log{
				UUID:    resp.Request.UUID,
				Level:   "debug",
				Message: fmt.Sprintf("HTTP %s => %s", resp.Request.Method, resp.Request.URL),
				Meta: map[string]interface{}{
					"duration": resp.Duration,
					"headers":  resp.Headers,
					"status":   resp.Status,
					"error":    resp.Error,
				},
			})
		})

	resultSignin, err := service.Login.Signin(entity.NewUUID(), "titi@mail.io", "Azerty1234.")
	if err != nil {
		fmt.Printf("Signin failed: %s\n", err)
		return
	}

	logger(entity.Log{
		UUID:    entity.NewUUID(),
		Level:   "debug",
		Message: "Signin result",
		Meta:    map[string]interface{}{"result": *resultSignin, "error": err},
	})

	resultMe, err := service.Login.Me(entity.NewUUID(), resultSignin.AccessToken)
	if err != nil {
		fmt.Printf("Me failed: %s\n", err)
		return
	}

	logger(entity.Log{
		UUID:    entity.NewUUID(),
		Level:   "debug",
		Message: "Me result",
		Meta:    map[string]interface{}{"result": *resultMe, "error": err},
	})

	err = service.Login.Logout(entity.NewUUID(), resultMe.AccessToken)
	if err != nil {
		fmt.Printf("Logout failed: %s\n", err)
		return
	}

	logger(entity.Log{
		UUID:    entity.NewUUID(),
		Level:   "debug",
		Message: "Logout result",
		Meta:    map[string]interface{}{"error": err},
	})
}
