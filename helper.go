package inspect

import (
	"strconv"
)

import (
	"github.com/bytepowered/fluxgo/pkg/flux"
)

// IsEmptyVars 判定列表参数是否任意一个为空值
func IsEmptyVars(vars []string) bool {
	if 0 == len(vars) {
		return true
	}
	for _, v := range vars {
		if "" == v {
			return true
		}
	}
	return false
}

type pagination struct {
	page     int
	pageSize int
	start    int
	end      int
}

func extraPageArgs(ctx flux.Context) pagination {
	page, pageSize := 1, 10
	if p, err := strconv.Atoi(ctx.FormVar("page")); err == nil {
		page = p
	}
	if ps, err := strconv.Atoi(ctx.FormVar("pageSize")); err == nil {
		pageSize = ps
	}
	// page, size 限制修正
	page = max(1, page)
	pageSize = min(max(1, pageSize), min(100, pageSize))
	// 索引计算
	start := max(0, page-1) * pageSize
	end := start + pageSize
	return pagination{
		page:     page,
		pageSize: pageSize,
		start:    start,
		end:      end,
	}
}

func limit(minV, maxV, v int) int {
	return max(minV, min(maxV, v))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
