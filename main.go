package s3server

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/cyd01/s3server/pkg/api"
	"github.com/cyd01/s3server/pkg/auth"
	"github.com/cyd01/s3server/pkg/logger"
	"github.com/cyd01/s3server/pkg/storage"
)

func Main(addr, dataDir string) {
	if !strings.Contains(addr, ":") {
		addr = ":" + addr
	}

	srv := &http.Server{
		Addr: addr,
	}

	log.Println("listening on " + addr + " and caching to " + dataDir)
	Start(srv, dataDir)
}

func Start(server *http.Server, dataDir string) {

	root := dataDir

	if err := os.MkdirAll(root, 0755); err != nil {
		log.Fatal(err)
	}

	fs := storage.NewFileSystem(root)

	handler := api.NewHandler(fs)

	authenticator := auth.NewAllowAllAuthenticator()
	/*
		store := auth.NewStaticCredentialStore(map[string]string{
			"test": "test",
		})
		authenticator := auth.NewSigV4Authenticator(
			store,
			"us-east-1",
			"s3",
		)
	*/

	middleware := auth.NewMiddleware(authenticator)

	server.Handler = logger.Log(true, middleware.Wrap(handler))

	log.Fatal(server.ListenAndServe())
}
