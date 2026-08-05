package service

import (
	"errors"
	"strings"
	"web3-wallet-exchange/global"
	"web3-wallet-exchange/model"

	"gorm.io/gorm"
)

// DepositListQuery 列表查询条件（对应前端筛选表单 + 分页条）。
// 和 Java 里 ListDepositReq / PageQuery 一类 DTO 同级。
type DepositListQuery struct {
	Page     int    // 从 1 开始
	PageSize int    // 默认 10，最大 50
	Status   string // 可选：pending / confirmed
	Asset    string // 可选：ETH
	TxHash   string // 可选：精确或前缀模糊（见实现）
}

// DepositListResult 列表响应体核心字段。
type DepositListResult struct {
	Items []model.Deposit
	Total int64
	Page  int
	Size  int
}

func ListDeposits(userId uint, q DepositListQuery) (*DepositListResult, error) {
	// —— 分页参数清洗（防前端乱传）——
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 10
	}
	if q.PageSize > 50 {
		q.PageSize = 50
	}

	// —— 动态 WHERE：先固定「只能看自己的」，再叠可选条件 ——
	// 注意：Count 和 Find 共用同一套条件，用 Session 避免互相污染。
	base := global.DB.Model(&model.Deposit{}).Where("user_id = ?", userId)

	if s := strings.TrimSpace(q.Status); s != "" {
		base = base.Where("status = ?", s)
	}
	if a := strings.TrimSpace(q.Asset); a != "" {
		base = base.Where("asset = ?", a)
	}
	if h := strings.ToLower(strings.TrimSpace(q.TxHash)); h != "" {
		// 输入完整 hash → 精确；输入片段 → 前缀匹配（演示用；生产可上索引策略）
		if len(h) >= 66 {
			base = base.Where("tx_hash = ?", h)
		} else {
			base = base.Where("tx_hash LIKE ?", h+"%")
		}
	}

	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, err
	}

	var items []model.Deposit
	offset := (q.Page - 1) * q.PageSize
	if err := base.Session(&gorm.Session{}).
		Order("id desc").
		Offset(offset).
		Limit(q.PageSize).
		Find(&items).Error; err != nil {
		return nil, err
	}
	if items == nil {
		items = []model.Deposit{}
	}

	return &DepositListResult{
		Items: items,
		Total: total,
		Page:  q.Page,
		Size:  q.PageSize,
	}, nil
}

// GetDepositByID 详情：必须校验归属当前用户（防越权看别人流水）。
func GetDepositByID(userId, id uint) (*model.Deposit, error) {
	var d model.Deposit
	err := global.DB.Where("id = ? AND user_id = ?", id, userId).First(&d).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &d, nil
}
