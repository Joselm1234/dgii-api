package contributors

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Contributor struct {
    bun.BaseModel `bun:"alias:c"`

    ID                    string    `json:"id" bun:",pk"`
    RNC                   string    `json:"rnc"`
    BusinessName          string    `json:"business_name"`
    CommercialName        string    `json:"commercial_name"`
    EconomicActivity      string    `json:"economic_activity"`
    StartDateOfOperations time.Time `json:"start_date_of_operations"`
    State                 string    `json:"state"`
    CreatedAt             time.Time `json:"createdAt" bun:",nullzero,notnull,default:current_timestamp"`
    UpdatedAt             time.Time `json:"updatedAt" bun:",nullzero,notnull,default:current_timestamp"`
}

var _ bun.BeforeAppendModelHook = (*Contributor)(nil)

func (c *Contributor) BeforeAppendModel(ctx context.Context, query bun.Query) error {
    now := time.Now().UTC()
    switch query.(type) {
    case *bun.InsertQuery:
        c.CreatedAt = now
    case *bun.UpdateQuery:
        c.UpdatedAt = now
    }
    return nil
}

func (c *Contributor) Save(ctx context.Context, db bun.IDB) error {
    if c.ID != "" {
        _, err := db.NewUpdate().Model(c).Where("id = ?", c.ID).Exec(ctx)
        if err != nil {
            log.Printf("error updating contributor: %v, id: %s", err, c.ID)
            return fmt.Errorf("error updating contributor: %w", err)
        }
        return nil
    }

    c.ID = uuid.NewString()
    _, err := db.NewInsert().Model(c).Exec(ctx)
    if err != nil {
        log.Printf("error inserting contributor: %v, id: %s", err, c.ID)
        return fmt.Errorf("error inserting contributor: %w", err)
    }
    return nil
}

func SelectContributorByRNC(ctx context.Context, db bun.IDB, rnc string) (*Contributor, error) {
    contributor := new(Contributor)
    err := db.NewSelect().
        Model(contributor).
        Where("rnc = ?", rnc).
        Scan(ctx)
    if err != nil {
        log.Printf("error selecting contributor by RNC: %v, rnc: %s", err, rnc)
        return nil, fmt.Errorf("error selecting contributor by RNC: %w", err)
    }
    return contributor, nil
}