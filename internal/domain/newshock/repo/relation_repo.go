// RelationRepo 关联关系数据仓库，管理 event_tickers 和 theme_tickers 两张中间表。
// 提供添加/删除关联、按一方查询另一方、清理某一方所有关联等方法。
package repo

import (
	"context"

	"gorm.io/gorm"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/newshock/dm"
)

type RelationRepo struct {
	db *gorm.DB
}

func NewRelationRepo(db *gorm.DB) *RelationRepo {
	return &RelationRepo{db: db}
}

// AddEventTicker 添加事件-股票关联（event_tickers 中间表）
func (r *RelationRepo) AddEventTicker(ctx context.Context, eventID, tickerID string) error {
	et := dm.EventTicker{EventID: eventID, TickerID: tickerID}
	if err := r.db.WithContext(ctx).Create(&et).Error; err != nil {
		return errors.WrapBizError(err, "关联事件股票失败")
	}
	return nil
}

// RemoveEventTicker 移除事件-股票关联
func (r *RelationRepo) RemoveEventTicker(ctx context.Context, eventID, tickerID string) error {
	if err := r.db.WithContext(ctx).Where("event_id = ? AND ticker_id = ?", eventID, tickerID).Delete(&dm.EventTicker{}).Error; err != nil {
		return errors.WrapBizError(err, "取消事件股票关联失败")
	}
	return nil
}

// GetTickersByEventID 获取事件关联的所有股票
func (r *RelationRepo) GetTickersByEventID(ctx context.Context, eventID string) ([]dm.EventTicker, error) {
	var relations []dm.EventTicker
	if err := r.db.WithContext(ctx).Where("event_id = ?", eventID).Find(&relations).Error; err != nil {
		return nil, errors.WrapBizError(err, "查询事件关联股票失败")
	}
	return relations, nil
}

// GetEventsByTickerID 获取股票关联的事件（按时间限制数量）
func (r *RelationRepo) GetEventsByTickerID(ctx context.Context, tickerID string, limit int) ([]dm.EventTicker, error) {
	var relations []dm.EventTicker
	if err := r.db.WithContext(ctx).Where("ticker_id = ?", tickerID).Limit(limit).Find(&relations).Error; err != nil {
		return nil, errors.WrapBizError(err, "查询股票关联事件失败")
	}
	return relations, nil
}

// AddThemeTicker 添加主题-股票关联（theme_tickers 中间表）
func (r *RelationRepo) AddThemeTicker(ctx context.Context, themeID, tickerID string) error {
	tt := dm.ThemeTicker{ThemeID: themeID, TickerID: tickerID}
	if err := r.db.WithContext(ctx).Create(&tt).Error; err != nil {
		return errors.WrapBizError(err, "关联主题股票失败")
	}
	return nil
}

// RemoveThemeTicker 移除主题-股票关联
func (r *RelationRepo) RemoveThemeTicker(ctx context.Context, themeID, tickerID string) error {
	if err := r.db.WithContext(ctx).Where("theme_id = ? AND ticker_id = ?", themeID, tickerID).Delete(&dm.ThemeTicker{}).Error; err != nil {
		return errors.WrapBizError(err, "取消主题股票关联失败")
	}
	return nil
}

// GetTickersByThemeID 获取主题关联的所有股票
func (r *RelationRepo) GetTickersByThemeID(ctx context.Context, themeID string) ([]dm.ThemeTicker, error) {
	var relations []dm.ThemeTicker
	if err := r.db.WithContext(ctx).Where("theme_id = ?", themeID).Find(&relations).Error; err != nil {
		return nil, errors.WrapBizError(err, "查询主题关联股票失败")
	}
	return relations, nil
}

// GetThemesByTickerID 获取股票关联的所有主题
func (r *RelationRepo) GetThemesByTickerID(ctx context.Context, tickerID string) ([]dm.ThemeTicker, error) {
	var relations []dm.ThemeTicker
	if err := r.db.WithContext(ctx).Where("ticker_id = ?", tickerID).Find(&relations).Error; err != nil {
		return nil, errors.WrapBizError(err, "查询股票关联主题失败")
	}
	return relations, nil
}

// ClearEventTickers 清除事件的所有股票关联（删除事件时调用）
func (r *RelationRepo) ClearEventTickers(ctx context.Context, eventID string) error {
	return r.db.WithContext(ctx).Where("event_id = ?", eventID).Delete(&dm.EventTicker{}).Error
}

// ClearThemeTickers 清除主题的所有股票关联（删除主题时调用）
func (r *RelationRepo) ClearThemeTickers(ctx context.Context, themeID string) error {
	return r.db.WithContext(ctx).Where("theme_id = ?", themeID).Delete(&dm.ThemeTicker{}).Error
}

// ClearThemeTickersByTicker 清除某股票的所有主题关联（删除股票时调用）
func (r *RelationRepo) ClearThemeTickersByTicker(ctx context.Context, tickerID string) error {
	return r.db.WithContext(ctx).Where("ticker_id = ?", tickerID).Delete(&dm.ThemeTicker{}).Error
}

// ClearEventTickersByTicker 清除某股票的所有事件关联（删除股票时调用）
func (r *RelationRepo) ClearEventTickersByTicker(ctx context.Context, tickerID string) error {
	return r.db.WithContext(ctx).Where("ticker_id = ?", tickerID).Delete(&dm.EventTicker{}).Error
}
