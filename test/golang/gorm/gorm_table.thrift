namespace go test

struct User {
    1: i64 id;
    2: string name (go.tag="json:\"json\" query:\"query\" form:\"form\" header:\"header\" goTag:\"taghhh\"");
} (gorm.table_name = "users_table")

struct Address {
    1: i64 id;
    2: string name (go.tag="json:\"json\" query:\"query\" form:\"form\" header:\"header\" goTag:\"taghhh\"");
}

// Only the first name will be used
struct Order{
    1: i64 id;
    2: string name (go.tag="json:\"json\" query:\"query\" form:\"form\" header:\"header\" goTag:\"taghhh\"");
} (
    gorm.table_name = "order",
    gorm.table_name = "order_table"
    )