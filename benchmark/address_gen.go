package benchmark

import (
	"fmt"
	"github.com/orbs-network/orbs-client-sdk-go/crypto/encoding"
	"github.com/orbs-network/orbs-client-sdk-go/orbs"
	"github.com/orbs-network/performance-benchmark/jsoncodec"
	"io/ioutil"
)

const KEY_COUNT = 100

func commandGenerateTestKeys(filename string) int {
	keys := make(map[string]*jsoncodec.Key)
	for i := 0; i < KEY_COUNT; i++ {
		account, err := orbs.CreateAccount()
		if err != nil {
			die("Could not create Orbs account.")
		}
		user := fmt.Sprintf("user%05d", i+1)
		keys[user] = &jsoncodec.Key{
			User:       user,
			PrivateKey: encoding.EncodeHex(account.PrivateKey),
			PublicKey:  encoding.EncodeHex(account.PublicKey),
			Address:    account.Address,
		}
	}

	bytes, err := jsoncodec.MarshalKeys(keys)
	if err != nil {
		die("Could not encode keys to json.\n\n%s", err.Error())
	}

	err = ioutil.WriteFile(filename, bytes, 0644)
	if err != nil {
		die("Could not write keys to file.\n\n%s", err.Error())
	}

	if !doesFileExist(filename) {
		die("File not found after write.")
	}

	log("%d new test keys written successfully to '%s'.\n", KEY_COUNT, filename)

	return KEY_COUNT
}

func getTestKeysFromFile(filename string) map[string]*jsoncodec.RawKey {
	if !doesFileExist(filename) {
		die("Input file %s does not exist", filename)
	}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		die("Could not open keys file '%s'.\n\n%s", filename, err.Error())
	}

	keys, err := jsoncodec.UnmarshalKeys(bytes)
	if err != nil {
		die("Failed parsing keys json file '%s'. Try deleting the key file to have it automatically recreated.\n\n%s", filename, err.Error())
	}

	all := make(map[string]*jsoncodec.RawKey)

	for _, key := range keys {
		id := key.User
		privateKey, err := encoding.DecodeHex(key.PrivateKey)
		if err != nil {
			die("Could not parse hex string '%s'. Try deleting the key file '%s' to have it automatically recreated.\n\n%s", privateKey, filename, err.Error())
		}

		publicKey, err := encoding.DecodeHex(key.PublicKey)
		if err != nil {
			die("Could not parse hex string '%s'. Try deleting the key file '%s' to have it automatically recreated.\n\n%s", publicKey, filename, err.Error())
		}

		address, err := encoding.DecodeHex(key.Address)
		if err != nil {
			die("Could not parse hex string '%s'. Try deleting the key file '%s' to have it automatically recreated.\n\n%s", address, filename, err.Error())
		}
		all[id] = &jsoncodec.RawKey{
			User:       id,
			PrivateKey: privateKey,
			PublicKey:  publicKey,
			Address:    address,
		}
	}
	fmt.Printf("Found %d addresses\n", len(all))
	return all
}
