package cli

import (
	"context"

	"github.com/lagz0ne/remmd/internal/core"
	"github.com/lagz0ne/remmd/internal/store"
)

func newWalker(links *store.LinkRepo) *core.GraphWalker {
	return core.NewGraphWalker(&linkQuerierAdapter{repo: links})
}

// linkQuerierAdapter adapts store.LinkRepo to core.LinkQuerier.
type linkQuerierAdapter struct {
	repo *store.LinkRepo
}

func (a *linkQuerierAdapter) LinksContainingSection(ctx context.Context, sectionID string) ([]*core.LinkInfo, error) {
	links, err := a.repo.LinksContainingSection(ctx, sectionID)
	if err != nil {
		return nil, err
	}
	result := make([]*core.LinkInfo, len(links))
	for i, l := range links {
		result[i] = &core.LinkInfo{
			ID:               l.ID,
			State:            string(l.State),
			RelationshipType: string(l.RelationshipType),
			LeftSectionIDs:   l.LeftSectionIDs,
			RightSectionIDs:  l.RightSectionIDs,
		}
	}
	return result, nil
}
