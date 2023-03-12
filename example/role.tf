#resource "athenz_role" "foo_role" {
#  domain = "home.nsegal.stage1"
#  name = "tf_role_test"
#  member {
#    name = "user.mshneorson"
#    expiration = "2023-12-29 23:59:59"
#  }
#  member {
#    name = "user.gbendor"
#  }
#  member {
#    name = "user.dguttman"
#    review = "2022-12-29 23:59:59"
#  }
#  member {
#    name = "user.relbaum"
#  }
#  tags = {
#    key1 = "val1,val2"
#    key2 = "val3,val4"
#  }
#  settings {
#    token_expiry_mins = 30
#    cert_expiry_mins = 25
#    user_expiry_days = 30
#    user_review_days = 7
#  }
#}

resource "athenz_role" "roleTest" {
  domain = "home.nsegal.stage1"
  name = "test_check"
  member {
    name = "user.nsegal"
#    expiration = "2022-12-29 23:59:59"
  }
  settings {
    token_expiry_mins = 5
    cert_expiry_mins = 10
#    user_expiry_days = 1
  }
  audit_ref="done by someone"
  tags = {
    key1 = "v1,v2"
    key2 = "v2,v3"
  }
}