// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

// Based on github.com/glebarez/sqlite v1.11.0.
// Original code copyright (c) 2013-NOW Jinzhu <wosmvp@gmail.com>,
// licensed under the MIT License. See LICENSE-glebarez-sqlite for the
// full MIT text. Inlined because the upstream package is unmaintained.

package sqlite

import "errors"

var ErrConstraintsNotImplemented = errors.New(
	"constraints not implemented on sqlite, consider using " +
		"DisableForeignKeyConstraintWhenMigrating, more details " +
		"https://github.com/go-gorm/gorm/wiki/GORM-V2-Release-Note-Draft" +
		"#all-new-migrator",
)
