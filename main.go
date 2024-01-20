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
func get_cred_config() Credentials {
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
	isProductionEnv := os.Getenv("PORT") != ""
	if isProductionEnv {
		log.Printf("--starting Production Development Environment--")
	} else {
		log.Printf("--starting Local Development Environment--")
		os.Setenv("AUTH_USERNAME", "admin")
		os.Setenv("AUTH_PASSWORD", "fake_password")
		// if err := godotenv.Load(".env"); err != nil {
		// 	log.Print("No .env file found in local development environment")
		// }
	}

	app := new(application)
	log.Print("starting server...")

	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultHTTPport
		log.Printf("defaulting to port %s", port)
	}
	log.Printf("listening on port %s", port)

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
		
	// Start HTTPS server
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
