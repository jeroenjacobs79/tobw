/*
 * Copyright (c) 2019 Jeroen Jacobs.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 2 as published by
 * the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package user

import (
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username     string
	Email        string
	passwordHash string
}

func (user *User) SetPassword(password string) (err error) {
	pwHash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return
	}
	user.passwordHash = string(pwHash)
	return
}

func (user *User) ValidatePassword(password string) (result bool) {
	err := bcrypt.CompareHashAndPassword([]byte(user.passwordHash), []byte(password))
	if err != nil {
		result = false
	} else {
		result = true
	}
	return
}

func (user *User) GetPasswordHash() (result string) {
	result = user.passwordHash
	return
}
