package contentapi

type Interface interface {
	CreateImageUrl(imgIdWithNamespace, permission string) (string, error)
}
