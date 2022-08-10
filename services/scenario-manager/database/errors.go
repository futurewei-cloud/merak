/*
MIT License
Copyright(c) 2022 Futurewei Cloud
    Permission is hereby granted,
    free of charge, to any person obtaining a copy of this software and associated documentation files(the "Software"), to deal in the Software without restriction,
    including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and / or sell copies of the Software, and to permit persons
    to whom the Software is furnished to do so, subject to the following conditions:
    The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
    THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
    FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
    WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

package database

// OperationError when cannot perform a given operation on database (SET, GET, DELETE)
type OperationError struct {
	operation string
}

func (err *OperationError) Error() string {
	return "Could not perform the " + err.operation + " operation."
}

// DownError when its not a redis.Nil response, in this case the database is down
type DownError struct{}

func (dbe *DownError) Error() string {
	return "Database is down."
}

// CreateDatabaseError when cannot perform set on database
type CreateDatabaseError struct{}

func (err *CreateDatabaseError) Error() string {
	return "Could not create Database."
}

// NotImplementedDatabaseError when user tries to create a not implemented database
type NotImplementedDatabaseError struct {
	databse string
}

func (err *NotImplementedDatabaseError) Error() string {
	return err.databse + "not implemented."
}
