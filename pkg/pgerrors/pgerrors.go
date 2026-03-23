package pgerrors

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/taiwoajasa245/torah_ai_backend/pkg/response"
)

// HandlePgError takes the error and writes the correct JSON response directly
// Usage in handler: pgerrors.Handle(w, err)
func HandlePgError(w http.ResponseWriter, err error) {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		response.Error(w, http.StatusInternalServerError, "Something went wrong, please try again", nil)
		return
	}

	log.Printf("[POSTGRES ERROR] Code: %s | Message: %s | Detail: %s", pgErr.Code, pgErr.Message, pgErr.Detail)

	code := pgErr.Code
	class := code[:2]

	switch code {

	// ── Class 08 — Connection Exceptions ─────────────────────────────────────
	case "08000":
		response.Error(w, http.StatusServiceUnavailable, "Database connection error", nil)
	case "08003":
		response.Error(w, http.StatusServiceUnavailable, "Database connection does not exist", nil)
	case "08006":
		response.Error(w, http.StatusServiceUnavailable, "Database connection failure", nil)
	case "08001":
		response.Error(w, http.StatusServiceUnavailable, "Unable to connect to database", nil)
	case "08004":
		response.Error(w, http.StatusServiceUnavailable, "Database connection rejected", nil)

	// ── Class 22 — Data Exceptions ────────────────────────────────────────────
	case "22001":
		response.Error(w, http.StatusBadRequest, "Value is too long for this field", nil)
	case "22003":
		response.Error(w, http.StatusBadRequest, "Number is out of the allowed range", nil)
	case "22007":
		response.Error(w, http.StatusBadRequest, "Invalid date format provided", nil)
	case "22008":
		response.Error(w, http.StatusBadRequest, "Date value is out of range", nil)
	case "22012":
		response.Error(w, http.StatusBadRequest, "Division by zero is not allowed", nil)
	case "22019":
		response.Error(w, http.StatusBadRequest, "Invalid escape character provided", nil)
	case "22025":
		response.Error(w, http.StatusBadRequest, "Invalid escape sequence provided", nil)
	case "22P02":
		response.Error(w, http.StatusBadRequest, "Invalid data type format provided", nil)
	case "22P03":
		response.Error(w, http.StatusBadRequest, "Invalid binary data format", nil)

	// ── Class 23 — Integrity Constraint Violations ────────────────────────────
	case "23000":
		response.Error(w, http.StatusConflict, "Data integrity constraint violated", nil)
	case "23001":
		response.Error(w, http.StatusConflict, "Operation would violate data restrictions", nil)
	case "23502":
		field := pgErr.ColumnName
		response.Error(w, http.StatusBadRequest, "Required field '"+field+"' cannot be empty", nil)
	case "23503":
		field := extractField(pgErr.Detail)
		response.Error(w, http.StatusUnprocessableEntity, "Related '"+field+"' resource does not exist", nil)
	case "23505":
		field := extractField(pgErr.Detail)
		response.Error(w, http.StatusConflict, "A record with this '"+field+"' already exists", nil)
	case "23514":
		response.Error(w, http.StatusBadRequest, "The value provided does not meet the required rules", nil)
	case "23P01":
		response.Error(w, http.StatusConflict, "Operation conflicts with existing data", nil)

	// ── Class 25 — Invalid Transaction State ──────────────────────────────────
	case "25000":
		response.Error(w, http.StatusInternalServerError, "Invalid transaction state", nil)
	case "25006":
		response.Error(w, http.StatusInternalServerError, "Cannot perform write operation in a read-only transaction", nil)
	case "25P02":
		response.Error(w, http.StatusInternalServerError, "Transaction was aborted, all operations are being skipped", nil)

	// ── Class 28 — Authorization Exceptions ───────────────────────────────────
	case "28000":
		response.Error(w, http.StatusUnauthorized, "Database authorization failed", nil)
	case "28P01":
		response.Error(w, http.StatusUnauthorized, "Invalid database credentials", nil)

	// ── Class 34 — Cursor Exceptions ──────────────────────────────────────────
	case "34000":
		response.Error(w, http.StatusInternalServerError, "Invalid database cursor", nil)

	// ── Class 40 — Transaction Rollback ───────────────────────────────────────
	case "40000":
		response.Error(w, http.StatusInternalServerError, "Transaction was rolled back", nil)
	case "40001":
		response.Error(w, http.StatusConflict, "Transaction conflict detected, please retry", nil)
	case "40002":
		response.Error(w, http.StatusInternalServerError, "Transaction integrity constraint violated", nil)
	case "40003":
		response.Error(w, http.StatusInternalServerError, "Transaction completion is uncertain", nil)
	case "40P01":
		response.Error(w, http.StatusConflict, "Deadlock detected, please retry your request", nil)

	// ── Class 42 — Syntax Error or Access Rule Violation ─────────────────────
	case "42000":
		response.Error(w, http.StatusInternalServerError, "Database syntax error occurred", nil)
	case "42501":
		response.Error(w, http.StatusForbidden, "Insufficient database permissions", nil)
	case "42601":
		response.Error(w, http.StatusInternalServerError, "Database syntax error occurred", nil)
	case "42602":
		response.Error(w, http.StatusInternalServerError, "Invalid database identifier", nil)
	case "42622":
		response.Error(w, http.StatusInternalServerError, "Database identifier is too long", nil)
	case "42701":
		response.Error(w, http.StatusConflict, "Duplicate column detected", nil)
	case "42702":
		response.Error(w, http.StatusInternalServerError, "Ambiguous column reference", nil)
	case "42703":
		response.Error(w, http.StatusInternalServerError, "Column does not exist in the database", nil)
	case "42704":
		response.Error(w, http.StatusInternalServerError, "Referenced object does not exist", nil)
	case "42710":
		response.Error(w, http.StatusConflict, "Duplicate database object detected", nil)
	case "42723":
		response.Error(w, http.StatusConflict, "Duplicate function definition detected", nil)
	case "42P01":
		response.Error(w, http.StatusInternalServerError, "Referenced table does not exist", nil)
	case "42P02":
		response.Error(w, http.StatusInternalServerError, "Parameter does not exist", nil)
	case "42P03":
		response.Error(w, http.StatusConflict, "Duplicate cursor detected", nil)
	case "42P04":
		response.Error(w, http.StatusConflict, "Duplicate database detected", nil)
	case "42P07":
		response.Error(w, http.StatusConflict, "Table already exists", nil)
	case "42P18":
		response.Error(w, http.StatusInternalServerError, "Indeterminate data type", nil)

	// ── Class 53 — Insufficient Resources ────────────────────────────────────
	case "53000":
		response.Error(w, http.StatusServiceUnavailable, "Database is out of resources", nil)
	case "53100":
		response.Error(w, http.StatusServiceUnavailable, "Database disk is full", nil)
	case "53200":
		response.Error(w, http.StatusServiceUnavailable, "Database is out of memory", nil)
	case "53300":
		response.Error(w, http.StatusServiceUnavailable, "Too many database connections, please try again later", nil)
	case "53400":
		response.Error(w, http.StatusServiceUnavailable, "Database configuration limit exceeded", nil)

	// ── Class 54 — Program Limit Exceeded ────────────────────────────────────
	case "54000":
		response.Error(w, http.StatusInternalServerError, "Database program limit exceeded", nil)
	case "54001":
		response.Error(w, http.StatusInternalServerError, "Database statement is too complex", nil)
	case "54011":
		response.Error(w, http.StatusInternalServerError, "Too many columns in the query", nil)
	case "54023":
		response.Error(w, http.StatusInternalServerError, "Too many arguments provided", nil)

	// ── Class 55 — Object Not In Prerequisite State ───────────────────────────
	case "55000":
		response.Error(w, http.StatusConflict, "Database object is not in the correct state", nil)
	case "55006":
		response.Error(w, http.StatusConflict, "Database object is currently in use", nil)
	case "55P03":
		response.Error(w, http.StatusConflict, "Resource is currently locked, please try again", nil)

	// ── Class 57 — Operator Intervention ─────────────────────────────────────
	case "57000":
		response.Error(w, http.StatusInternalServerError, "Database operation was interrupted", nil)
	case "57014":
		response.Error(w, http.StatusRequestTimeout, "Request took too long and was cancelled", nil)
	case "57P01":
		response.Error(w, http.StatusServiceUnavailable, "Database is shutting down", nil)
	case "57P02":
		response.Error(w, http.StatusServiceUnavailable, "Database crashed unexpectedly", nil)
	case "57P03":
		response.Error(w, http.StatusServiceUnavailable, "Database is not accepting connections right now", nil)
	case "57P04":
		response.Error(w, http.StatusServiceUnavailable, "Database session was terminated", nil)

	// ── Class 58 — System Error ───────────────────────────────────────────────
	case "58000":
		response.Error(w, http.StatusInternalServerError, "Unexpected system error occurred", nil)
	case "58030":
		response.Error(w, http.StatusInternalServerError, "Database IO error occurred", nil)

	// ── Fallback by class ─────────────────────────────────────────────────────
	default:
		switch class {
		case "08":
			response.Error(w, http.StatusServiceUnavailable, "Database connection problem, please try again", nil)
		case "22":
			response.Error(w, http.StatusBadRequest, "Invalid data provided", nil)
		case "23":
			response.Error(w, http.StatusConflict, "Data conflict, please check your input", nil)
		case "25", "40":
			response.Error(w, http.StatusInternalServerError, "transaction error, please try again", nil)
		case "42":
			response.Error(w, http.StatusInternalServerError, "Something went wrong, please try again", nil)
		case "53", "54":
			response.Error(w, http.StatusServiceUnavailable, "Service is temporarily unavailable", nil)
		case "57", "58":
			response.Error(w, http.StatusServiceUnavailable, "Database is temporarily unavailable", nil)
		default:
			response.Error(w, http.StatusInternalServerError, "Something went wrong, please try again", nil)
		}
	}
}

// extractField pulls the column name out of postgres detail message
// e.g. "Key (email)=(test@test.com) already exists" → "email"
func extractField(detail string) string {
	start := strings.Index(detail, "(")
	end := strings.Index(detail, ")")
	if start != -1 && end != -1 && end > start {
		return detail[start+1 : end]
	}
	return "value"
}
