## 2. Dependencias (go.mod)

```go
module github.com/yourorg/smtp-client

go 1.21

require (
    golang.org/x/oauth2 v0.15.0
    google.golang.org/api v0.152.0
    github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.4.0
    github.com/sirupsen/logrus v1.9.3
    github.com/google/uuid v1.4.0
    golang.org/x/sync v0.5.0
    gopkg.in/yaml.v3 v3.0.1
    github.com/joho/godotenv v1.5.1
    github.com/spf13/viper v1.17.0
    github.com/go-playground/validator/v10 v10.16.0
)

