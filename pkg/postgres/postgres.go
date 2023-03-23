package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/coreweave/ncore-api/pkg/database"
	"github.com/coreweave/ncore-api/pkg/ipxe"
	"github.com/coreweave/ncore-api/pkg/payloads"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB handles database communication with PostgreSQL.
type DB struct {
	Postgres *pgxpool.Pool
}

// txCtx key.
type txCtx struct{}

// connCtx key.
type connCtx struct{}

// conn returns a PostgreSQL transaction if one exists.
// If not, returns a connection if a connection has been acquired by calling WithAcquire.
// Otherwise, it returns *pgxpool.Pool which acquires the connection and closes it immediately after a SQL command is executed.
func (db *DB) conn(ctx context.Context) database.PGXQuerier {
	if tx, ok := ctx.Value(txCtx{}).(pgx.Tx); ok && tx != nil {
		return tx
	}
	if res, ok := ctx.Value(connCtx{}).(*pgxpool.Conn); ok && res != nil {
		return res
	}
	return db.Postgres
}

var _ ipxe.DB = (*DB)(nil)     // Check if methods expected by ipxe.DB are implemented correctly.
var _ payloads.DB = (*DB)(nil) // Check if methods expected by payloads.DB are implemented correctly.

type nodePayload struct {
	PayloadId        string
	PayloadDirectory string
	MacAddress       string
}

func (np *nodePayload) dto() *payloads.NodePayload {
	return &payloads.NodePayload{
		PayloadId:        np.PayloadId,
		PayloadDirectory: np.PayloadDirectory,
		MacAddress:       np.MacAddress,
	}
}

type ipxeDbConfig struct {
	ImageName   string
	ImageBucket string
	ImageTag    string
	ImageType   string
  ImageCmdline string
}

func (ic *ipxeDbConfig) dto() *ipxe.IpxeDbConfig {
	return &ipxe.IpxeDbConfig{
		ImageName:   ic.ImageName,
		ImageBucket: ic.ImageBucket,
		ImageTag:    ic.ImageTag,
		ImageType:   ic.ImageType,
    ImageCmdline: ic.ImageCmdline,
	}
}

// GetPayload returns a Payload.
// TODO: Convert to return a list of all payloads for a mac
// TODO: In image always take index 0 of list
func (db *DB) GetNodePayload(ctx context.Context, macAddress string) (*payloads.NodePayload, error) {
	var np []nodePayload
	np_sql := fmt.Sprintf(`
			SELECT
				node_payloads.payload_id,
				payloads.payload_directory,
				node_payloads.mac_address
			FROM "node_payloads"
			JOIN payloads on (node_payloads.payload_id = payloads.payload_id)
			WHERE mac_address like '%s' LIMIT 1
	`, macAddress) // #nosec G201
	np_rows, err := db.conn(ctx).Query(ctx, np_sql)
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return nil, err
	}
	if err == nil {
		np, err = pgx.CollectRows(np_rows, pgx.RowToStructByPos[nodePayload])
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		log.Printf("cannot get node_payload from database: %v\n", err)
		return nil, errors.New("cannot get node_payload from database")
	}
	if len(np) == 0 {
		return nil, nil
	}

	return np[0].dto(), nil
}

func (db *DB) AddDefaultNodePayload(ctx context.Context, config *payloads.NodePayloadDb) (*payloads.NodePayloadDb, error) {
	var npd *payloads.NodePayloadDb
	const npd_sql = `
    INSERT INTO node_payloads (
      payload_id,
      mac_address
    )
    VALUES (
        $1,
        $2
    );
	`
	switch _, err := db.conn(ctx).Exec(ctx, npd_sql,
    config.PayloadId,
    config.MacAddress,
  ); {
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return nil, err
	case err != nil:
		log.Printf("Error - AddDefaultNodePayload: %v - %v\n", err, config)
		return nil, fmt.Errorf(`cannot create payload for config: %v`, config)
	}

	npd = &payloads.NodePayloadDb{
		PayloadId:   config.PayloadId,
		MacAddress:   config.MacAddress,
	}
	return npd, nil
}

// UpdateNodePayload updates the PayloadId for mac_address.
func (db *DB) UpdateNodePayload(ctx context.Context, config *payloads.NodePayloadDb) (*payloads.NodePayloadDb, error) {
	var npd *payloads.NodePayloadDb
  log.Printf("UpdateNodePayload: %v\n", *config)
	const npd_sql = `
    UPDATE node_payloads
    SET
        payload_id=$1
    WHERE
        mac_address like $2
  `
	switch _, err := db.conn(ctx).Exec(ctx, npd_sql,
    config.PayloadId,
    config.MacAddress,
  ); {
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return nil, err
	case err != nil:
		log.Printf("Error - UpdateNodePayload: %v - %v\n", err, config)
		return nil, fmt.Errorf(`cannot update payload for config: %v`, *config)
	}

	npd = &payloads.NodePayloadDb{
		PayloadId:   config.PayloadId,
		MacAddress:   config.MacAddress,
	}
	return npd, nil
}

// GetPayloadParameters returns an interface{} for a payloadId.
func (db *DB) GetPayloadParameters(ctx context.Context, payloadId string) (interface{}, error) {
	// https://faraday.ai/blog/how-to-aggregate-jsonb-in-postgres/
	sql := fmt.Sprintf(`
		with params as (select distinct
					payload_schemas.parameter_name,
					payload_parameters.parameter_value
			from payload_schemas
			join payloads on (
					payload_schemas.payload_schema_id = payloads.payload_schema_id
			)
			join payload_parameters on (
					payload_schemas.parameter_name = payload_parameters.parameter_name
			) where payload_parameters.payload_id = '%s')
			select jsonb_object_agg(jsonb_build_object(params.parameter_name, params.parameter_value)) from params
	`, payloadId)
	pp_rows, err := db.conn(ctx).Query(ctx, sql)
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return nil, err
	}
	if err == nil {
		var result interface{}
		for pp_rows.Next() {
			values, _ := pp_rows.Values()
			for _, v := range values {
				result = v
			}
		}
		if result == nil {
			fmt.Printf("result is nil")
			return nil, nil
		}
		return result, err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		log.Printf("cannot get payload_parameters from database: %v\n", err)
		return nil, errors.New("cannot get payload_parameters from database")
	}
	return nil, nil
}

// GetIpxe returns an IpxeConfig for a macAddress.
// TODO: https://github.com/uber-go/zap
func (db *DB) GetIpxeDbConfig(ctx context.Context, macAddress string) (*ipxe.IpxeDbConfig, error) {
	var ic []ipxeDbConfig
	sql := fmt.Sprintf(`
    SELECT
        images.image_name,
        images.image_bucket,
        images.image_tag,
        images.image_type,
        images.image_cmdline
    FROM images
    JOIN node_images on (
      node_images.image_tag = images.image_tag
    ) AND (
      node_images.image_type = images.image_type
    )
      WHERE node_images.mac_address = '%s';
	`, macAddress)
	ic_rows, err := db.conn(ctx).Query(ctx, sql)
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return nil, err
	}
	if err == nil {
		ic, err = pgx.CollectRows(ic_rows, pgx.RowToStructByPos[ipxeDbConfig])
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, pgx.ErrNoRows
	}
	if err != nil {
		log.Printf("cannot get ipxe from database: %v\n", err)
		return nil, errors.New("cannot get ipxe from database")
	}
	if len(ic) == 0 {
		return nil, errors.New("no image found in database")
	}
	return ic[0].dto(), nil
}

// CreateIpxeImage inserts an IpxeDbConfig into ipxe.images.
func (db *DB) CreateIpxeImage(ctx context.Context, config *ipxe.IpxeDbConfig) (*ipxe.IpxeConfig, error) {
	var ic *ipxe.IpxeConfig
	const sql = `
    INSERT INTO images (
        image_name,
        image_bucket,
        image_tag,
        image_type,
        image_cmdline
    )
    VALUES (
        $1,
        $2,
        $3,
        $4,
        $5
    );
	`
	switch _, err := db.conn(ctx).Exec(ctx, sql,
		config.ImageName,
		config.ImageBucket,
		config.ImageTag,
		config.ImageType,
		config.ImageCmdline,
	); {
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return nil, err
	case err != nil:
		if sqlErr := db.createIpxeImagePgError(err); sqlErr != nil {
			return nil, sqlErr
		}
		log.Printf("cannot create image: %v\n", err)
		return nil, errors.New("cannot create image")
	}
	ic = &ipxe.IpxeConfig{
		ImageName:   config.ImageName,
		ImageBucket: config.ImageBucket,
		ImageTag:    config.ImageTag,
		ImageType:   config.ImageType,
		ImageCmdline:   config.ImageCmdline,
	}
	return ic, nil
}

// DeleteIpxeImage deletes an entry in ipxe.images matching image_tag and image_type.
func (db *DB) DeleteIpxeImage(ctx context.Context, config *ipxe.IpxeDbDeleteConfig) (*ipxe.IpxeDbConfig, error) {
	var idc *ipxeDbConfig
	sql := fmt.Sprintf(`
    DELETE FROM images
    WHERE
        image_tag = '%s'
        AND
        image_type = '%s'
    RETURNING (
        image_name,
        image_bucket,
        image_tag,
        image_type
    )
	`,
    config.ImageTag,
    config.ImageType,
  )
  row := db.conn(ctx).QueryRow(ctx, sql)
	switch err := row.Scan(&idc); {
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return nil, err
	case err != nil:
		if sqlErr := db.deleteIpxeImagePgError(err); sqlErr != nil {
			return nil, sqlErr
		}
		log.Printf("cannot delete image: %v\n", err)
		return nil, errors.New("cannot delete image")
	default:
    return idc.dto(), nil
  }
}

func (db *DB) createIpxeImagePgError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}
	if pgErr.Code == pgerrcode.UniqueViolation {
		return errors.New("UniqueViolation - (image_tag, image_type) exists")
	}
	if pgErr.Code == pgerrcode.CheckViolation {
		switch pgErr.ConstraintName {
		case "image_bucket":
			return errors.New("invalid image_bucket")
		case "image_name":
			return errors.New("invalid image_name")
		case "image_tag":
			return errors.New("invalid image_tag")
		case "image_type":
			return errors.New("invalid image_type")
    case "image_cmdline":
      return errors.New("invalid image_cmdline")
    }
	}
	return nil
}

func (db *DB) deleteIpxeImagePgError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}
	return errors.New("Error deleting (image_tag, image_type): " + pgErr.Code)
}
