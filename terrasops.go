package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"flag"
	"github.com/getsops/sops/v3"
	"github.com/getsops/sops/v3/aes"
	"github.com/getsops/sops/v3/cmd/sops/common"
	"github.com/getsops/sops/v3/cmd/sops/formats"
	"github.com/getsops/sops/v3/config"
	"github.com/getsops/sops/v3/decrypt"
	"github.com/getsops/sops/v3/keyservice"
	"github.com/getsops/sops/v3/stores/json"
	"github.com/getsops/sops/v3/version"
	"github.com/labstack/echo/v4"
)

var rulesFile = flag.String("rules", ".sops.yaml", "path to sops creation rules")
var stateFile = flag.String("state", "tfstate.sops", "path to encrypted terraform state")
var listenAddr = flag.String("listen", "127.0.0.1:5000", "address to listen on")

func rerr(c echo.Context, err error) error {
	return c.String(http.StatusInternalServerError, err.Error())
}

func main() {
	flag.Parse()

	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		ciphertext, err := ioutil.ReadFile(*stateFile)
		if err != nil {
			return rerr(c, err)
		}
		state, err := decrypt.DataWithFormat(ciphertext, formats.Json)
		if err != nil {
			return rerr(c, err)
		}
		return c.Blob(http.StatusOK, "application/json", state)
	})

	e.POST("/", func(c echo.Context) error {
		plaintext, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			return rerr(c, err)
		}
		branches, err := (&json.Store{}).LoadPlainFile(plaintext)
		if err != nil {
			return rerr(c, err)
		}
		rules, err := config.LoadCreationRuleForFile(*rulesFile, *stateFile, map[string]*string{})
		if err != nil {
			return rerr(c, err)
		}
		tree := sops.Tree{
			Branches: branches,
			Metadata: sops.Metadata{
				Version:         version.Version,
				KeyGroups:       rules.KeyGroups,
				ShamirThreshold: rules.ShamirThreshold,
			},
		}
		keys, errs := tree.GenerateDataKeyWithKeyServices([]keyservice.KeyServiceClient{keyservice.NewLocalClient()})
		if len(errs) > 0 {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("%s", errs))
		}
		if common.EncryptTree(common.EncryptTreeOpts{
			DataKey: keys,
			Tree:    &tree,
			Cipher:  aes.NewCipher(),
		}) != nil {
			return rerr(c, err)
		}
		ciphertext, err := (&json.Store{}).EmitEncryptedFile(tree)
		if err != nil {
			return rerr(c, err)
		}
		if ioutil.WriteFile(*stateFile, ciphertext, 0o0600) != nil {
			return rerr(c, err)
		}
		return c.String(http.StatusOK, "state updated")
	})

	e.Logger.Fatal(e.Start(*listenAddr))
}
