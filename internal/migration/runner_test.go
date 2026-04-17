package migration

import (
	"reflect"
	"testing"
)

func TestSplitOracleStatementsSplitsPlainSQL(t *testing.T) {
	input := `
CREATE USER WINTHOR IDENTIFIED BY secret;
GRANT CREATE SESSION TO WINTHOR;
`

	got := splitOracleStatements(input)
	want := []string{
		"CREATE USER WINTHOR IDENTIFIED BY secret",
		"GRANT CREATE SESSION TO WINTHOR",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("splitOracleStatements() = %#v, want %#v", got, want)
	}
}

func TestSplitOracleStatementsKeepsAnonymousBlockTogether(t *testing.T) {
	input := `
BEGIN
    EXECUTE IMMEDIATE 'DROP USER WINTHOR CASCADE';
EXCEPTION
    WHEN OTHERS THEN
        IF SQLCODE != -1918 THEN
            RAISE;
        END IF;
END;
/
`

	got := splitOracleStatements(input)
	want := []string{
		"BEGIN\n    EXECUTE IMMEDIATE 'DROP USER WINTHOR CASCADE';\nEXCEPTION\n    WHEN OTHERS THEN\n        IF SQLCODE != -1918 THEN\n            RAISE;\n        END IF;\nEND;",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("splitOracleStatements() = %#v, want %#v", got, want)
	}
}

func TestSplitOracleStatementsKeepsCreateOrReplaceTriggerTogether(t *testing.T) {
	input := `
CREATE OR REPLACE TRIGGER trg_test
BEFORE INSERT ON pcs
FOR EACH ROW
BEGIN
    :NEW.created_at := SYSTIMESTAMP;
END;
/

ALTER TRIGGER trg_test ENABLE;
`

	got := splitOracleStatements(input)
	want := []string{
		"CREATE OR REPLACE TRIGGER trg_test\nBEFORE INSERT ON pcs\nFOR EACH ROW\nBEGIN\n    :NEW.created_at := SYSTIMESTAMP;\nEND;",
		"ALTER TRIGGER trg_test ENABLE",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("splitOracleStatements() = %#v, want %#v", got, want)
	}
}
