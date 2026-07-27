package cli

//go:generate sh -c "cd .. && rm -rf cmd/rpgtextbox && go run github.com/arran4/go-subcommand/cmd/gosubc@v0.0.17 generate --dir . && go run golang.org/x/tools/cmd/goimports@latest -w cmd/rpgtextbox/ && rm -f cmd/rpgtextbox/samples_test.go cmd/rpgtextbox/skill_test.go cmd/rpgtextbox/install_test.go cmd/rpgtextbox/update_test.go cmd/rpgtextbox/remove_test.go cmd/rpgtextbox/list_test.go cmd/rpgtextbox/inspect_test.go"
