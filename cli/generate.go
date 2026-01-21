package cli

//go:generate sh -c "rm -rf ../cmd/rpgtextbox && cd .. && go run github.com/arran4/go-subcommand/cmd/gosubc generate --dir . && sed -i '/^\t\"\"$/d' cmd/rpgtextbox/*.go"
