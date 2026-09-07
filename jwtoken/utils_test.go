package jwtoken

// Функция для создания PEM файла с ключом - для удобства в виде теста сделан
/*func TestMakePEMPrivateKey(t *testing.T) {
	pkey, err := GenerateKey()
	assert.Nil(t, err)
	privateKeyPEM, publicKeyPEM := CreatePEM(pkey)

	privateEMFile, _ := os.Create("private_key.pem")
	pem.Encode(privateEMFile, privateKeyPEM)

	publicEMFile, _ := os.Create("public_key.pem")
	pem.Encode(publicEMFile, publicKeyPEM)
}

func TestLoadPrivateEMFile(t *testing.T) {
	bytes, err := os.ReadFile("pkey.pem")
	assert.Nil(t, err)
	ppem, _ := pem.Decode(bytes)
	_, err = LoadPrivateEM(ppem)
	assert.Nil(t, err)
}*/
