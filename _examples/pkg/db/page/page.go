package page

type Info[E any] struct {
	PageNo   int64 `json:"pageNo"`
	PageSize int64 `json:"pageSize"`
	Count    int64 `json:"count"`
	MaxPage  int64 `json:"maxPage"`
	Records  []E   `json:"records"`
}

// NewPageInfo 创建分页信息
// @param pageNo: 页码
// @param pageSize: 每页大小
// @param count: 总记录数
// @param maxPage: 最大页码
// @param records: 记录
// @return *Info[E]: 分页信息
func NewPageInfo[E any](pageNo, pageSize, count, maxPage int64, records []E) *Info[E] {
	return &Info[E]{
		PageNo:   pageNo,
		PageSize: pageSize,
		Count:    count,
		MaxPage:  maxPage,
		Records:  records,
	}
}

// GetPageNo 获取页码
// @return int64: 页码
func (e *Info[E]) GetPageNo() int64 {
	return e.PageNo
}

// GetPageSize 获取每页大小
// @return int64: 每页大小
func (e *Info[E]) GetPageSize() int64 {
	return e.PageSize
}

// GetCount 获取总记录数
// @return int64: 总记录数
func (e *Info[E]) GetCount() int64 {
	return e.Count
}

// GetMaxPage 获取最大页码
// @return int64: 最大页码
func (e *Info[E]) GetMaxPage() int64 {
	return e.MaxPage
}

// GetRecords 获取记录
// @return []E: 记录
func (e *Info[E]) GetRecords() []E {
	return e.Records
}

// SetPageNo 设置页码
// @param pageNo: 页码
func (e *Info[E]) SetPageNo(pageNo int64) {
	e.PageNo = pageNo
}

// SetPageSize 设置每页大小
// @param pageSize: 每页大小
func (e *Info[E]) SetPageSize(pageSize int64) {
	e.PageSize = pageSize
}

// SetCount 设置总记录数
// @param count: 总记录数
func (e *Info[E]) SetCount(count int64) {
	e.Count = count
}

// SetMaxPage 设置最大页码
// @param maxPage: 最大页码
func (e *Info[E]) SetMaxPage(maxPage int64) {
	e.MaxPage = maxPage
}

// SetRecords 设置记录
// @param records: 记录
func (e *Info[E]) SetRecords(records []E) {
	e.Records = records
}
