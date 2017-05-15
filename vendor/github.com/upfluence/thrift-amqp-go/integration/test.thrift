exception TestException {}

service Foo {
  i64 add(1: i64 x, 2: i64 y),
  void ping() throws (1:TestException exc),
  oneway void yolo(),
}
