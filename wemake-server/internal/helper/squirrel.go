package helper

import (
	sq "github.com/Masterminds/squirrel"
)

func ApplyPagination(builder sq.SelectBuilder, offset int, limit int) sq.SelectBuilder {
	if limit > 0 {
		builder = builder.Limit(uint64(limit))
	}
	if offset >= 0 {
		builder = builder.Offset(uint64(offset))
	}
	return builder
}

func ApplySort(builder sq.SelectBuilder, column string, direction string, allowedFields []string) (sq.SelectBuilder, error) {
	if column == "" {
		return builder, nil
	}

	allowed := false
	for _, field := range allowedFields {
		if field == column {
			allowed = true
			break
		}
	}
	if !allowed {
		return builder, ErrInvalidSortColumn
	}

	if direction != "ASC" && direction != "DESC" {
		direction = "ASC"
	}

	return builder.OrderBy(column + " " + direction), nil
}

func BuildEqFilter(filters map[string]interface{}) sq.Eq {
	eq := sq.Eq{}
	if filters == nil {
		return eq
	}
	for key, value := range filters {
		if value != nil {
			eq[key] = value
		}
	}
	return eq
}

func BuildAndFilter(conditions ...sq.Sqlizer) sq.And {
	andConds := sq.And{}
	for _, cond := range conditions {
		if cond != nil {
			andConds = append(andConds, cond)
		}
	}
	return andConds
}

func BuildOrFilter(conditions ...sq.Sqlizer) sq.Or {
	orConds := sq.Or{}
	for _, cond := range conditions {
		if cond != nil {
			orConds = append(orConds, cond)
		}
	}
	return orConds
}

func BuildInFilter(column string, values []interface{}) sq.Sqlizer {
	if len(values) == 0 {
		return nil
	}
	return sq.Eq{column: values}
}

func BuildRangeFilter(column string, min interface{}, max interface{}) sq.And {
	conds := sq.And{}
	if min != nil {
		conds = append(conds, sq.GtOrEq{column: min})
	}
	if max != nil {
		conds = append(conds, sq.LtOrEq{column: max})
	}
	return conds
}

func BuildLikeFilter(column string, value string) sq.Sqlizer {
	if value == "" {
		return nil
	}
	return sq.Like{column: "%" + value + "%"}
}

func CountQuery(builder sq.SelectBuilder) (string, []interface{}, error) {
	countBuilder := sq.Select("COUNT(*)").FromSelect(builder, "t")
	return countBuilder.ToSql()
}

func IntersectOperator(conds ...sq.Sqlizer) sq.And {
	return BuildAndFilter(conds...)
}

func UnionOperator(conds ...sq.Sqlizer) sq.Or {
	return BuildOrFilter(conds...)
}

func SafePageParams(pageNum int, pageSize int, minPageNum int, minPageSize int, maxPageSize int) (int, int) {
	return MaxIntQuery(pageNum, minPageNum), ClampInt(pageSize, minPageSize, maxPageSize)
}

func OffsetFromPage(pageNum int, pageSize int) int {
	return CalculateOffset(pageNum, pageSize)
}

func PageFromOffset(offset int, pageSize int) int {
	if pageSize <= 0 {
		return 1
	}
	return (offset / pageSize) + 1
}

func PrepareSQL(builder sq.SelectBuilder) sq.SelectBuilder {
	return builder
}
