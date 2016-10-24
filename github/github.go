/*
   Copyright (C) 2016 The BlameWarrior Authors.

   This file is a part of BlameWarrior service.

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/
package github

import "strings"

// SplitRepositoryName splits full GitHub repository name into owner and name parts.
func SplitRepositoryName(fullName string) (owner, repo string) {
	sep := strings.IndexByte(fullName, '/')
	if sep <= 0 || sep == len(fullName)-1 {
		return "", ""
	}

	return fullName[0:sep], fullName[sep+1:]
}
