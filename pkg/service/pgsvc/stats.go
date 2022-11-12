package pgsvc

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/shopspring/decimal"
	"optrispace.com/work/pkg/db/pgdao"
	"optrispace.com/work/pkg/model"
)

// StatsSvc service
type StatsSvc struct {
	db *sql.DB
}

// NewStats returns new Stats
func NewStats(db *sql.DB) *StatsSvc {
	return &StatsSvc{
		db: db,
	}
}

// Stats implements service.Stats interface
func (s *StatsSvc) Stats(ctx context.Context) (*model.Stats, error) {
	var result *model.Stats
	return result, doWithQueries(ctx, s.db, defaultRoTxOpts, func(queries *pgdao.Queries) error {
		result = &model.Stats{
			RegistrationsByDate: map[string]int{},
		}

		registrationsByDate, err := queries.StatRegistrationsByDate(ctx)
		if err != nil {
			return fmt.Errorf("unable to StatRegistrationsByDate: %w", err)
		}

		for _, r := range registrationsByDate {
			result.RegistrationsByDate[r.Day.Format("2006-01-02")] = int(r.Registrations)
		}

		registrationsCount, err := queries.StatsGetRegistrationsCount(ctx)
		if err != nil {
			return fmt.Errorf("unable to StatsGetRegistrationsCount: %w", err)
		}
		result.TotalRegistrations = registrationsCount

		openedJobs, err := queries.StatsGetOpenedJobsCount(ctx)
		if err != nil {
			return fmt.Errorf("unable to StatsGetOpenedJobsCount: %w", err)
		}

		result.OpenedJobs = openedJobs

		contractsCount, err := queries.StatsGetContractsCount(ctx)
		if err != nil {
			return fmt.Errorf("unable to StatsGetContractsCount: %w", err)
		}

		result.TotalContracts = contractsCount

		contractsVolume, err := queries.StatsGetContractsVolume(ctx)
		if err != nil {
			return fmt.Errorf("unable to StatsGetContractsVolume: %w", err)
		}

		result.TotalTransactionsVolume = decimal.RequireFromString(contractsVolume)

		return nil
	})
}
