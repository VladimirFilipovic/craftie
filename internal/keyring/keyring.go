package keyring

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

const (
	serviceDest     = "org.freedesktop.secrets"
	servicePath     = "/org/freedesktop/secrets"
	serviceIface    = "org.freedesktop.Secret.Service"
	collectionIface = "org.freedesktop.Secret.Collection"
	itemIface       = "org.freedesktop.Secret.Item"
	collection      = "/org/freedesktop/secrets/aliases/default"
)

type KeyringSession struct {
	conn    *dbus.Conn
	session dbus.ObjectPath
}

type dbusSecret struct {
	Session     dbus.ObjectPath
	Parameters  []byte
	Value       []byte
	ContentType string
}

func Open() (*KeyringSession, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, fmt.Errorf("dbus connect: %w", err)
	}

	obj := conn.Object(serviceDest, dbus.ObjectPath(servicePath))
	var algo dbus.Variant
	var session dbus.ObjectPath
	err = obj.Call(serviceIface+".OpenSession", 0, "plain", dbus.MakeVariant("")).Store(&algo, &session)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("open session: %w", err)
	}

	return &KeyringSession{conn: conn, session: session}, nil
}

func (k *KeyringSession) Close() error {
	return k.conn.Close()
}

func (k *KeyringSession) Get(service, key string) (string, error) {
	obj := k.conn.Object(serviceDest, dbus.ObjectPath(servicePath))
	attrs := map[string]string{
		"application": service,
		"username":    key,
	}
	var unlocked, locked []dbus.ObjectPath
	if err := obj.Call(serviceIface+".SearchItems", 0, attrs).Store(&unlocked, &locked); err != nil {
		return "", fmt.Errorf("search items: %w", err)
	}
	items := unlocked
	if len(items) == 0 {
		return "", fmt.Errorf("secret not found: %s/%s", service, key)
	}

	var secret dbusSecret
	// we use replace = true when adding secret so we can be sure there are no duplicate
	// thus its safe to take first item
	secrConObj := k.conn.Object(serviceDest, dbus.ObjectPath(items[0]))
	err := secrConObj.Call(itemIface+".GetSecret", 0, k.session).Store(&secret)

	if err != nil {
		return "", fmt.Errorf("get secret: %w", err)
	}
	return string(secret.Value), nil
}

func (k *KeyringSession) Set(service, key, value string) error {
	collection := k.conn.Object(serviceDest, collection)

	props := map[string]dbus.Variant{
		"org.freedesktop.Secret.Item.Label": dbus.MakeVariant(service + "/" + key),
		"org.freedesktop.Secret.Item.Attributes": dbus.MakeVariant(map[string]string{
			"application": service,
			"username":    key,
		}),
	}

	secret := dbusSecret{
		Session:     k.session,
		Parameters:  []byte{},
		Value:       []byte(value),
		ContentType: "text/plain",
	}

	var item dbus.ObjectPath
	var prompt dbus.ObjectPath
	// replace=true to update existing items
	err := collection.Call(collectionIface+".CreateItem", 0, props, secret, true).Store(&item, &prompt)
	if err != nil {
		return fmt.Errorf("create item: %w", err)
	}
	return nil
}

func openSession(conn *dbus.Conn) (dbus.ObjectPath, error) {
	obj := conn.Object(serviceDest, dbus.ObjectPath(servicePath))
	var algo dbus.Variant
	var session dbus.ObjectPath
	err := obj.Call(serviceIface+".OpenSession", 0, "plain", dbus.MakeVariant("")).Store(&algo, &session)
	if err != nil {
		return "", fmt.Errorf("open session: %w", err)
	}
	return session, nil
}
