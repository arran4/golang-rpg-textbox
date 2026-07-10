package cli

//go:generate sh -c "rm -rf ../cmd/rpgtextbox && cd .. && go run github.com/arran4/go-subcommand/cmd/gosubc@v0.0.17 generate --dir . && sed -i '/^\t\"\"$/d' cmd/rpgtextbox/*.go && sed -i 's/c\\.FlagSet\\.PrintDefaults()/c.PrintDefaults()/g' cmd/rpgtextbox/*.go && sed -i 's/c\\.FlagSet\\.Parse(args)/c.Parse(args)/g' cmd/rpgtextbox/*.go && sed -i 's/c\\.FlagSet\\.Args()/c.Args()/g' cmd/rpgtextbox/*.go"
