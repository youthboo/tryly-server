package me

import (
	"database/sql"
	"errors"

	"github.com/yourusername/wemake/internal/domain"
	merepo "github.com/yourusername/wemake/internal/repository/me"
)

var ErrRFQOrderNotFound = errors.New("rfq not found")

type RFQOrdersService struct {
	repo *merepo.RFQOrdersRepository
}

func NewRFQOrdersService(repo *merepo.RFQOrdersRepository) *RFQOrdersService {
	return &RFQOrdersService{repo: repo}
}

func (s *RFQOrdersService) List(userID int64) ([]domain.MeRFQOrderSummary, error) {
	rows, err := s.repo.ListSummaries(userID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.MeRFQOrderSummary, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapRFQOrderSummary(row))
	}
	return out, nil
}

func (s *RFQOrdersService) GetDetail(userID, rfqID int64) (*domain.MeRFQOrderDetail, error) {
	rfqRow, err := s.repo.GetDetailRFQ(userID, rfqID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRFQOrderNotFound
		}
		return nil, err
	}

	quoteRows, err := s.repo.ListQuotations(rfqID)
	if err != nil {
		return nil, err
	}
	quotations := make([]domain.MeRFQOrderQuotation, 0, len(quoteRows))
	for _, row := range quoteRows {
		quotations = append(quotations, mapRFQOrderQuotation(row))
	}

	var order *domain.MeRFQOrderOrder
	if ordRow, e := s.repo.GetLatestOrder(userID, rfqID); e == nil {
		order = &domain.MeRFQOrderOrder{
			OrderID:     ordRow.OrderID,
			OrderStatus: ordRow.OrderStatus,
			TotalAmount: ordRow.TotalAmount,
			CreatedAt:   ordRow.CreatedAt,
		}
	} else if !errors.Is(e, sql.ErrNoRows) {
		return nil, e
	}

	detail := mapRFQOrderDetail(*rfqRow, quotations, order)
	return &detail, nil
}

func mapRFQOrderSummary(row merepo.RFQOrderSummaryRow) domain.MeRFQOrderSummary {
	item := domain.MeRFQOrderSummary{
		RFQID:          row.RFQID,
		Title:          row.Title,
		RequestKind:    row.RequestKind,
		Status:         row.Status,
		CreatedAt:      row.CreatedAt,
		QuotationCount: row.QuotationCount,
	}
	if row.CategoryName.Valid {
		v := row.CategoryName.String
		item.CategoryName = &v
	}
	if row.TargetPrice.Valid {
		v := row.TargetPrice.Float64
		item.TargetPrice = &v
	}
	if row.OrderID.Valid {
		v := row.OrderID.Int64
		item.OrderID = &v
	}
	if row.OrderStatus.Valid {
		v := row.OrderStatus.String
		item.OrderStatus = &v
	}
	if row.TotalAmount.Valid {
		v := row.TotalAmount.Float64
		item.TotalAmount = &v
	}
	if row.FactoryID.Valid {
		v := row.FactoryID.Int64
		item.FactoryID = &v
	}
	if row.FactoryName.Valid {
		v := row.FactoryName.String
		item.FactoryName = &v
	}
	if row.EstimatedDelivery.Valid {
		v := row.EstimatedDelivery.String
		item.EstimatedDelivery = &v
	}
	if row.OrderCreatedAt.Valid {
		v := row.OrderCreatedAt.String
		item.OrderCreatedAt = &v
	}
	return item
}

func mapRFQOrderQuotation(row merepo.RFQOrderQuotationRow) domain.MeRFQOrderQuotation {
	return domain.MeRFQOrderQuotation{
		QuoteID:     row.QuoteID,
		FactoryName: row.FactoryName,
		GrandTotal:  row.GrandTotal,
		Status:      row.Status,
		CreatedAt:   row.CreatedAt,
	}
}

func mapRFQOrderDetail(row merepo.RFQOrderDetailRFQRow, quotations []domain.MeRFQOrderQuotation, order *domain.MeRFQOrderOrder) domain.MeRFQOrderDetail {
	detail := domain.MeRFQOrderDetail{
		RFQID:       row.RFQID,
		Title:       row.Title,
		RequestKind: row.RequestKind,
		Status:      row.Status,
		CreatedAt:   row.CreatedAt,
		Quantity:    row.Quantity,
		Quotations:  quotations,
		Order:       order,
	}
	if row.CategoryName.Valid {
		v := row.CategoryName.String
		detail.CategoryName = &v
	}
	if row.Details.Valid {
		v := row.Details.String
		detail.Details = &v
	}
	if row.TargetPrice.Valid {
		v := row.TargetPrice.Float64
		detail.TargetPrice = &v
	}
	if row.TargetLeadTimeDays.Valid {
		v := int(row.TargetLeadTimeDays.Int64)
		detail.TargetLeadTimeDays = &v
	}
	if detail.Quotations == nil {
		detail.Quotations = []domain.MeRFQOrderQuotation{}
	}
	return detail
}
