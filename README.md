# Bootsmann #

Rewrite all config file types on docker startup.

Example of template sources:

```
  # Declare global section
  # All parameters declared here will be
  # available in all sections if not overwritten
  [$global]

  # Default template boundaries, separated by space
  $template = #{{ }}

  # Name of variable and value
  GLOBVAR = 55


  # Process a config file
  [example/cfg1.conf]
  # Overwrite default template boundaries
  $template = %[[ ]]
  PARAM = "simple string"

  [example/cfg2.ini]
  VALUE = "12345"

  # Process a shell name pattern
  # All files in this directory will be processed
  [example/dir/*]
  VAR = 1024
  $template = !<< >>
```
