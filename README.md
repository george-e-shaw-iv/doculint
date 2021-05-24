# doculint

A Go linter that focuses on proper commenting.

## Features

- Validates package names are not mixed case and do not contain `-` or `_`.
- Validates that packages have a comment beginning with `Package <package name>` in a file with the same name as the
package.
- Validates that all function declarations have a comment beginning with the name of the function.
- Validates that all constant and type blocks have a comment associated with them.
- Validates that all constants and type declarations have comments associated with them.
- Validates that literals are not used in conditional expressions found in if statements.
