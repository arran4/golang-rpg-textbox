package cli

//go:generate sh -c "rm -rf ../cmd/rpgtextbox && cd .. && go run github.com/arran4/go-subcommand/cmd/gosubc@v0.0.17 generate --dir . && sed -i '/^\t\"\"$/d' cmd/rpgtextbox/*.go"
