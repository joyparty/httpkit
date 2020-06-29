package httpkit

import (
	"math"
)

// Pagination 数据库分页计算
type Pagination struct {
	First    int `json:"first"`
	Last     int `json:"last"`
	Previous int `json:"previous"`
	Current  int `json:"current"`
	Next     int `json:"next"`
	Size     int `json:"size"`
	Items    int `json:"items"`
}

// NewPagination 计算分页页码
func NewPagination(current, size, items int) Pagination {
	if current <= 0 {
		current = 1
	}

	p := Pagination{
		First:   1,
		Last:    1,
		Current: current,
	}

	if size > 0 {
		p.Size = size
	}
	if items > 0 {
		p.Items = items
	}

	if items > 0 && size > 0 {
		p.Last = int(math.Ceil(float64(p.Items) / float64(p.Size)))
	}

	if p.Current < p.First {
		p.Current = p.First
	} else if p.Current > p.Last {
		p.Current = p.Last
	}

	if p.Current > p.First {
		p.Previous = p.Current - 1
	}
	if p.Current < p.Last {
		p.Next = p.Current + 1
	}

	return p
}

// Limit 数据库查询LIMIT值
func (p Pagination) Limit() int {
	return p.Size
}

// Offset 数据库查询OFFSET值
func (p Pagination) Offset() int {
	return (p.Current - 1) * p.Size
}
