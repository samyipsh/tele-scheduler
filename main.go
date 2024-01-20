// Sample run-helloworld is a minimal Cloud Run service.
package main

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
)

type application struct {
    auth struct {
        username string
        password string
    }
}

type Credentials struct {
	Username string `json:"AUTH_USERNAME"`
	Password string `json:"AUTH_PASSWORD"`
}

// const keyUsername = "username"
const keyServerAddr = "serverAddr"
// const portGoogleCloudRunDefault = "9090"
const defaultHTTPport = "8080"


func getRoot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	serverAddr := ctx.Value(keyServerAddr).(string)

	// read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("%s: error reading body: %s\n", serverAddr, err)
	}

	fmt.Printf("%s: got / request. body:\n%s\n", serverAddr, body)
	fmt.Fprintf(w, "This is my website!\n") // responseWriter implements io.Writer
}

func getHello(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	serverAddr := ctx.Value(keyServerAddr).(string)

	fmt.Printf("%s: got /hello request\n", serverAddr)
	w.Write([]byte("Hello, HTTP!\n")) // more declarative to just use w.Write
}

// Retrieve Cloud credentials stored in SecretManager or default to environment variables.
// Returns: Credentials struct with non-empty username and password
// TODO: get secrets even in local dev: https://cloud.google.com/secret-manager/docs/samples/secretmanager-get-secret?hl=en#secretmanager_get_secret-go 
func get_cred_config() Credentials {
	// credentialsJSON := os.Getenv("CLOUD_SQL_CREDENTIALS_SECRET")
	// if credentialsJSON != "" {
	// 	log.Printf("Using CLOUD_SQL_CREDENTIALS_SECRET")
	// 	var credentials Credentials
	// 	err := json.Unmarshal([]byte(credentialsJSON), &credentials)
	// 	if err != nil {
	// 		log.Fatalf("Unable to parse CLOUD_SQL_CREDENTIALS_SECRET: %v", err)
	// 	}
	// 	return credentials
	// }

	log.Printf("Using environment variables to get credentials")
	username := os.Getenv("AUTH_USERNAME")
	if username == "" {
		log.Fatal("basic auth username must be provided")
	}
	password := os.Getenv("AUTH_PASSWORD")
	if password == "" {	
		log.Fatal("basic auth password must be provided")
	}
	return Credentials{Username: username, Password: password}
}

func main() {
		app := new(application)
        log.Print("starting server...")

		isProductionEnv := os.Getenv("PORT") != ""
		log.Printf("isProductionEnv: %t", isProductionEnv)

		// Determine port for HTTP service.
		port := os.Getenv("PORT")
		if port == "" {
			port = defaultHTTPport
			log.Printf("defaulting to port %s", port)
		}

		credentials := get_cred_config()
		app.auth.username = credentials.Username
		app.auth.password = credentials.Password

		// create your own http.Handler
		mux := http.NewServeMux()
		mux.HandleFunc("/", getRoot)
		mux.HandleFunc("/hello", app.basicAuth(getHello))

		ctx := context.Background()
		server := &http.Server{
			Addr: ":" + port,
			Handler: mux,
			BaseContext: func(l net.Listener) context.Context {
				ctx = context.WithValue(ctx, keyServerAddr, l.Addr().String())
				return ctx
			},
		}

        log.Printf("listening on port %s", port)
		if isProductionEnv {
			log.Printf("starting production server...\n")
		} else {
			log.Printf("starting local dev server...\n")
		}
			
		// Start HTTP server.
		err := server.ListenAndServeTLS("./localhost.pem", "./localhost-key.pem");
		if errors.Is(err, http.ErrServerClosed) {
			log.Printf("server closed\n")
		} else if err != nil {
			log.Fatal(err)
		}
}

func (app *application) basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))
			expectedUsernameHash := sha256.Sum256([]byte(app.auth.username))
			expectedPasswordHash := sha256.Sum256([]byte(app.auth.password))

			usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
			passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

// package main

// import (
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"os"
// )

// func main() {
//         log.Print("starting server...")
//         http.HandleFunc("/", handler)

//         // Determine port for HTTP service.
//         port := os.Getenv("PORT")
//         if port == "" {
//                 port = "8080"
//                 log.Printf("defaulting to port %s", port)
//         }

//         // Start HTTP server.
//         log.Printf("listening on port %s", port)
//         if err := http.ListenAndServe(":"+port, nil); err != nil {
//                 log.Fatal(err)
//         }
// }

// func handler(w http.ResponseWriter, r *http.Request) {
//         name := os.Getenv("NAME")
//         if name == "" {
//                 name = "World"
//         }
//         fmt.Fprintf(w, "Hello %s!\n", name)
// }