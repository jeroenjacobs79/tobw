/*
 * Copyright (c) 2019 Head In Cloud BVBA.
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
 *
 *
 */

package ansiterm

import "io"

type AnsiTerminal struct {
	io.ReadWriteCloser
}

func CreateAnsiTerminal(device io.ReadWriteCloser) (*AnsiTerminal){
	term := AnsiTerminal{
		ReadWriteCloser: device,
	}
	return &term
}