# Bootsmann #

Rewrite all config file types on docker startup.

Example of template sources:

```
  [$global]
  $template = #{{ }}
  GLOBVAR = 55


  [example/cfg1.conf]
  $template = %[[ ]]
  PARAM = "simple string"

  [example/cfg2.ini]
  VALUE = "12345"

  [example/dir/*]
  VAR = 1024
  $template = !<< >>

```
