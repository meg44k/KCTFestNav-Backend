package repository

import (
	"context"
	"database/sql"

	"github.com/meg44k/KCTFestNav-Backend/internal/database"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

type stageRepository struct {
	sqlDB *sql.DB
	db    *database.Queries
}

func NewStageRepository(db *sql.DB) domain.StageRepository {
	return &stageRepository{sqlDB: db, db: database.New(db)}
}

func toSection(s database.StageSection) *domain.StageSection {
	return &domain.StageSection{
		ID:        int(s.ID),
		Name:      s.Name,
		Location:  s.Location,
		SortOrder: int(s.SortOrder),
		Blocks:    []*domain.StageBlock{},
	}
}

func toBlock(b database.StageBlock) *domain.StageBlock {
	return &domain.StageBlock{
		ID:           int(b.ID),
		SectionID:    int(b.SectionID),
		StartTime:    b.StartTime,
		EndTime:      b.EndTime,
		CurrentOrder: int(b.CurrentOrder),
		Performers:   []*domain.Performer{},
	}
}

func toPerformer(p database.Performer) *domain.Performer {
	return &domain.Performer{
		ID:           int(p.ID),
		BlockID:      int(p.BlockID),
		Name:         p.Name,
		Detail:       p.Detail,
		ThumbnailURL: p.ThumbnailUrl,
		PerformOrder: int(p.PerformOrder),
	}
}

// 3 つの表をそれぞれ並べて取り、入れ子にする(件数は多くて数百なので全部読む)
func (r *stageRepository) GetSchedule(ctx context.Context) ([]*domain.StageSection, error) {
	dbSections, err := r.db.ListStageSections(ctx)
	if err != nil {
		return nil, err
	}
	dbBlocks, err := r.db.ListStageBlocks(ctx)
	if err != nil {
		return nil, err
	}
	dbPerformers, err := r.db.ListPerformers(ctx)
	if err != nil {
		return nil, err
	}

	sections := make([]*domain.StageSection, len(dbSections))
	sectionByID := map[int]*domain.StageSection{}
	for i, s := range dbSections {
		sections[i] = toSection(s)
		sectionByID[sections[i].ID] = sections[i]
	}
	blockByID := map[int]*domain.StageBlock{}
	for _, b := range dbBlocks {
		block := toBlock(b)
		blockByID[block.ID] = block
		if s, ok := sectionByID[block.SectionID]; ok {
			s.Blocks = append(s.Blocks, block)
		}
	}
	for _, p := range dbPerformers {
		performer := toPerformer(p)
		if b, ok := blockByID[performer.BlockID]; ok {
			b.Performers = append(b.Performers, performer)
		}
	}
	return sections, nil
}

func (r *stageRepository) GetSection(ctx context.Context, id int) (*domain.StageSection, error) {
	s, err := r.db.GetStageSection(ctx, int32(id))
	if err != nil {
		return nil, err
	}
	return toSection(s), nil
}

func (r *stageRepository) GetBlock(ctx context.Context, id int) (*domain.StageBlock, error) {
	b, err := r.db.GetStageBlock(ctx, int32(id))
	if err != nil {
		return nil, err
	}
	block := toBlock(b)
	performers, err := r.db.ListPerformersInBlock(ctx, b.ID)
	if err != nil {
		return nil, err
	}
	for _, p := range performers {
		block.Performers = append(block.Performers, toPerformer(p))
	}
	return block, nil
}

func (r *stageRepository) GetPerformer(ctx context.Context, id int) (*domain.Performer, error) {
	p, err := r.db.GetPerformer(ctx, int32(id))
	if err != nil {
		return nil, err
	}
	return toPerformer(p), nil
}

func insertedID(res sql.Result, err error) (int, error) {
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

func (r *stageRepository) CreateSection(ctx context.Context, s *domain.StageSection) (int, error) {
	return insertedID(r.db.CreateStageSection(ctx, database.CreateStageSectionParams{
		Name:      s.Name,
		Location:  s.Location,
		SortOrder: int32(s.SortOrder),
	}))
}

// 値が変わらない UPDATE は影響行数が 0 になるため、存在は先に Get で確かめる
func (r *stageRepository) UpdateSection(ctx context.Context, s *domain.StageSection) error {
	if _, err := r.db.GetStageSection(ctx, int32(s.ID)); err != nil {
		return err
	}
	return r.db.UpdateStageSection(ctx, database.UpdateStageSectionParams{
		Name:      s.Name,
		Location:  s.Location,
		SortOrder: int32(s.SortOrder),
		ID:        int32(s.ID),
	})
}

// 中のブロックと出演者は外部キーの ON DELETE CASCADE で消える
func (r *stageRepository) DeleteSection(ctx context.Context, id int) error {
	if _, err := r.db.GetStageSection(ctx, int32(id)); err != nil {
		return err
	}
	return r.db.DeleteStageSection(ctx, int32(id))
}

func (r *stageRepository) CreateBlock(ctx context.Context, b *domain.StageBlock) (int, error) {
	// 親が無いときは外部キーのエラーではなく 404 にしたいので先に確かめる
	if _, err := r.db.GetStageSection(ctx, int32(b.SectionID)); err != nil {
		return 0, err
	}
	return insertedID(r.db.CreateStageBlock(ctx, database.CreateStageBlockParams{
		SectionID: int32(b.SectionID),
		StartTime: b.StartTime,
		EndTime:   b.EndTime,
	}))
}

// 時刻だけを変える。current_order(学生会が進めた位置)は保つ
func (r *stageRepository) UpdateBlock(ctx context.Context, b *domain.StageBlock) error {
	if _, err := r.db.GetStageBlock(ctx, int32(b.ID)); err != nil {
		return err
	}
	return r.db.UpdateStageBlock(ctx, database.UpdateStageBlockParams{
		StartTime: b.StartTime,
		EndTime:   b.EndTime,
		ID:        int32(b.ID),
	})
}

func (r *stageRepository) DeleteBlock(ctx context.Context, id int) error {
	if _, err := r.db.GetStageBlock(ctx, int32(id)); err != nil {
		return err
	}
	return r.db.DeleteStageBlock(ctx, int32(id))
}

// ブロックの行を FOR UPDATE で押さえ、出演順の採番や入れ替えが同時に走っても重ならないようにする
func (r *stageRepository) inBlockTx(ctx context.Context, blockID int, fn func(q *database.Queries) error) error {
	tx, err := r.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var locked int
	if err := tx.QueryRowContext(ctx, "SELECT id FROM stage_blocks WHERE id = ? FOR UPDATE", blockID).Scan(&locked); err != nil {
		return err
	}
	if err := fn(r.db.WithTx(tx)); err != nil {
		return err
	}
	return tx.Commit()
}

// 出演順はそのブロックの最後(最大 + 1)
func (r *stageRepository) CreatePerformer(ctx context.Context, p *domain.Performer) (int, error) {
	var id int
	err := r.inBlockTx(ctx, p.BlockID, func(q *database.Queries) error {
		next, err := q.NextPerformOrder(ctx, int32(p.BlockID))
		if err != nil {
			return err
		}
		id, err = insertedID(q.CreatePerformer(ctx, database.CreatePerformerParams{
			BlockID:      int32(p.BlockID),
			Name:         p.Name,
			Detail:       p.Detail,
			ThumbnailUrl: p.ThumbnailURL,
			PerformOrder: int32(next),
		}))
		return err
	})
	return id, err
}

// 名前・紹介・写真だけを変える。出演順は MovePerformer で変える
func (r *stageRepository) UpdatePerformer(ctx context.Context, p *domain.Performer) error {
	if _, err := r.db.GetPerformer(ctx, int32(p.ID)); err != nil {
		return err
	}
	return r.db.UpdatePerformer(ctx, database.UpdatePerformerParams{
		Name:         p.Name,
		Detail:       p.Detail,
		ThumbnailUrl: p.ThumbnailURL,
		ID:           int32(p.ID),
	})
}

func (r *stageRepository) DeletePerformer(ctx context.Context, id int) error {
	if _, err := r.db.GetPerformer(ctx, int32(id)); err != nil {
		return err
	}
	return r.db.DeletePerformer(ctx, int32(id))
}

// 同じブロックの隣の出演者と出演順を入れ替える。端なら何もしない
func (r *stageRepository) MovePerformer(ctx context.Context, id int, d domain.MoveDirection) error {
	target, err := r.db.GetPerformer(ctx, int32(id))
	if err != nil {
		return err
	}
	return r.inBlockTx(ctx, int(target.BlockID), func(q *database.Queries) error {
		performers, err := q.ListPerformersInBlock(ctx, target.BlockID)
		if err != nil {
			return err
		}
		for i, p := range performers {
			if p.ID != target.ID {
				continue
			}
			j := i - 1
			if d == domain.MoveDown {
				j = i + 1
			}
			if j < 0 || j >= len(performers) {
				return nil
			}
			other := performers[j]
			if err := q.SetPerformerOrder(ctx, database.SetPerformerOrderParams{PerformOrder: other.PerformOrder, ID: p.ID}); err != nil {
				return err
			}
			return q.SetPerformerOrder(ctx, database.SetPerformerOrderParams{PerformOrder: p.PerformOrder, ID: other.ID})
		}
		return sql.ErrNoRows // 取ってから消された
	})
}

func (r *stageRepository) AdvanceBlock(ctx context.Context, id int) error {
	if _, err := r.db.GetStageBlock(ctx, int32(id)); err != nil {
		return err
	}
	return r.db.AdvanceBlock(ctx, database.AdvanceBlockParams{ID: int32(id)})
}

func (r *stageRepository) RewindBlock(ctx context.Context, id int) error {
	if _, err := r.db.GetStageBlock(ctx, int32(id)); err != nil {
		return err
	}
	return r.db.RewindBlock(ctx, int32(id))
}
