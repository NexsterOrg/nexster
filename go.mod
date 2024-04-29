module github.com/NamalSanjaya/nexster

go 1.20

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.8.0
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.4.0
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.2.0
	github.com/arangodb/go-driver v1.6.0
	github.com/euroteltr/rbac v1.0.9
	github.com/go-mail/mail v2.3.1+incompatible
	github.com/go-playground/validator/v10 v10.14.1
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/golang-jwt/jwt/v5 v5.0.0
	github.com/google/uuid v1.3.1
	github.com/hbollon/go-edlib v1.6.0
	github.com/julienschmidt/httprouter v1.3.0
	github.com/labstack/gommon v0.4.1
	github.com/rs/cors v1.9.0
	golang.org/x/crypto v0.14.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.3.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.1.1 // indirect
	github.com/arangodb/go-velocypack v0.0.0-20200318135517-5af53c29c67e // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/net v0.15.0 // indirect
	golang.org/x/sys v0.14.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/mail.v2 v2.3.1 // indirect
)

replace github.com/NamalSanjaya/nexster/pkgs/models/userInsight => ./pkgs/models/userInsight

replace github.com/NamalSanjaya/nexster/pkgs/models/userInsightOf => ./pkgs/models/userInsightOf
