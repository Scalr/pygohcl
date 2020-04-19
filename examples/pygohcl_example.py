import pygohcl

print(pygohcl.Parse('variable "test" { default = "abc" }'))
