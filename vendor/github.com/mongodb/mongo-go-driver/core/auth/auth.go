// Copyright (C) MongoDB, Inc. 2017-present.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package auth

import (
	"context"
	"fmt"

	"github.com/mongodb/mongo-go-driver/core/addr"
	"github.com/mongodb/mongo-go-driver/core/command"
	"github.com/mongodb/mongo-go-driver/core/connection"
	"github.com/mongodb/mongo-go-driver/core/description"
	"github.com/mongodb/mongo-go-driver/core/wiremessage"
)

// AuthenticatorFactory constructs an authenticator.
type AuthenticatorFactory func(cred *Cred) (Authenticator, error)

var authFactories = make(map[string]AuthenticatorFactory)

func init() {
	RegisterAuthenticatorFactory("", newDefaultAuthenticator)
	RegisterAuthenticatorFactory(SCRAMSHA1, newScramSHA1Authenticator)
	RegisterAuthenticatorFactory(MONGODBCR, newMongoDBCRAuthenticator)
	RegisterAuthenticatorFactory(PLAIN, newPlainAuthenticator)
	RegisterAuthenticatorFactory(GSSAPI, newGSSAPIAuthenticator)
	RegisterAuthenticatorFactory(MongoDBX509, newMongoDBX509Authenticator)
}

// CreateAuthenticator creates an authenticator.
func CreateAuthenticator(name string, cred *Cred) (Authenticator, error) {
	if f, ok := authFactories[name]; ok {
		return f(cred)
	}

	return nil, fmt.Errorf("unknown authenticator: %s", name)
}

// RegisterAuthenticatorFactory registers the authenticator factory.
func RegisterAuthenticatorFactory(name string, factory AuthenticatorFactory) {
	authFactories[name] = factory
}

// // Opener returns a connection opener that will open and authenticate the connection.
// func Opener(opener conn.Opener, authenticator Authenticator) conn.Opener {
// 	return func(ctx context.Context, addr model.Addr, opts ...conn.Option) (conn.Connection, error) {
// 		return NewConnection(ctx, authenticator, opener, addr, opts...)
// 	}
// }
//
// // NewConnection opens a connection and authenticates it.
// func NewConnection(ctx context.Context, authenticator Authenticator, opener conn.Opener, addr model.Addr, opts ...conn.Option) (conn.Connection, error) {
// 	conn, err := opener(ctx, addr, opts...)
// 	if err != nil {
// 		if conn != nil {
// 			// Ignore any error that occurs since we're already returning a different one.
// 			_ = conn.Close()
// 		}
// 		return nil, err
// 	}
//
// 	err = authenticator.Auth(ctx, conn)
// 	if err != nil {
// 		// Ignore any error that occurs since we're already returning a different one.
// 		_ = conn.Close()
// 		return nil, err
// 	}
//
// 	return conn, nil
// }

// Configurer creates a connection configurer for the given authenticator.
//
// TODO(skriptble): Fully implement this once this package is moved over to the new connection type.
// func Configurer(configurer connection.Configurer, authenticator Authenticator) connection.Configurer {
// 	return connection.ConfigurerFunc(func(ctx context.Context, conn connection.Connection) (connection.Connection, error) {
// 		err := authenticator.Auth(ctx, conn)
// 		if err != nil {
// 			conn.Close()
// 			return nil, err
// 		}
// 		if configurer == nil {
// 			return conn, nil
// 		}
// 		return configurer.Configure(ctx, conn)
// 	})
// }

// Handshaker creates a connection handshaker for the given authenticator. The
// handshaker will handle calling isMaster and buildInfo.
func Handshaker(appName string, h connection.Handshaker, authenticator Authenticator) connection.Handshaker {
	return connection.HandshakerFunc(func(ctx context.Context, address addr.Addr, rw wiremessage.ReadWriter) (description.Server, error) {
		desc, err := (&command.Handshake{Client: command.ClientDoc(appName)}).Handshake(ctx, address, rw)
		if err != nil {
			return description.Server{}, err
		}

		err = authenticator.Auth(ctx, desc, rw)
		if err != nil {
			return description.Server{}, err
		}
		if h == nil {
			return desc, nil
		}
		return h.Handshake(ctx, address, rw)
	})
}

// Authenticator handles authenticating a connection.
type Authenticator interface {
	// Auth authenticates the connection.
	Auth(context.Context, description.Server, wiremessage.ReadWriter) error
}

func newError(err error, mech string) error {
	return &Error{
		message: fmt.Sprintf("unable to authenticate using mechanism \"%s\"", mech),
		inner:   err,
	}
}

// Error is an error that occurred during authentication.
type Error struct {
	message string
	inner   error
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.message, e.inner)
}

// Inner returns the wrapped error.
func (e *Error) Inner() error {
	return e.inner
}

// Message returns the message.
func (e *Error) Message() string {
	return e.message
}
