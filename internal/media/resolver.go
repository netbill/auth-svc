package media

type Resolver struct {
	baseUrl string
}

func NewResolver(baseUrl string) *Resolver {
	return &Resolver{
		baseUrl: baseUrl,
	}
}

func (r *Resolver) Resolve(key string) string {
	return r.baseUrl + "/" + key
}
