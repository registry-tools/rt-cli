resource "null_resource" "foo" {
  triggers = {
    hello = "world"
  }
}
