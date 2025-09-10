package requests

type NodeListRequest struct {
	PageNo   int  `json:"page_no" yaml:"page_no" form:"pageNo" binding:"required"`
	PageSize int  `json:"page_size" yaml:"page_size" form:"pageSize" binding:"required"`
	IsPage   bool `json:"is_page" yaml:"is_page" form:"isPage" binding:"-"`
}

func (n *NodeListRequest) IsLegal() bool {
	if n.PageSize <= 0 || n.PageNo <= 0 {
		return false
	}
	return true
}
