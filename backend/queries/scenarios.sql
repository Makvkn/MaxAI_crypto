-- name: CreateScenarioCalculation :one
INSERT INTO scenario_calculations (
    user_id, wallet_id, type, asset_id, change_pct,
    baseline_portfolio_value_usd, baseline_asset_value_usd, baseline_asset_allocation_pct,
    projected_portfolio_value_usd, projected_asset_value_usd, asset_impact_usd,
    portfolio_change_usd, portfolio_change_pct,
    data_quality, calculation_version
)
VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8,
    $9, $10, $11,
    $12, $13,
    $14, $15
)
RETURNING
    id, user_id, wallet_id, type, asset_id, change_pct,
    baseline_portfolio_value_usd, baseline_asset_value_usd, baseline_asset_allocation_pct,
    projected_portfolio_value_usd, projected_asset_value_usd, asset_impact_usd,
    portfolio_change_usd, portfolio_change_pct,
    data_quality, calculation_version, created_at;

-- name: GetScenarioCalculationByID :one
SELECT
    id, user_id, wallet_id, type, asset_id, change_pct,
    baseline_portfolio_value_usd, baseline_asset_value_usd, baseline_asset_allocation_pct,
    projected_portfolio_value_usd, projected_asset_value_usd, asset_impact_usd,
    portfolio_change_usd, portfolio_change_pct,
    data_quality, calculation_version, created_at
FROM scenario_calculations
WHERE id = $1;
