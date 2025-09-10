package binding

// BodyLegal request body
type BodyLegal interface {
	// IsLegal 请求是否合法
	IsLegal() bool
}
