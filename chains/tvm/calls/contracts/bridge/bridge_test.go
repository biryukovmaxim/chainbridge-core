package bridge

//
//const (
//	bridgeHex             = "TPH9cWgafMHhGmzL3ccaWX5gF7e8kbicZr"
//	network               = "grpc.shasta.trongrid.io:50051"
//	expectedHandlerBase58 = "TBq9Rc5mPtq7tLHBxnHUXGkuaEDxrKX3ya"
//	resourceHex           = "0x6d792d746f6b656e000000000000000000000000000000000000000000000000"
//)
//
//func TestBridge(t *testing.T) {
//	client := tronClient.NewGrpcClient(network)
//
//	err := client.Start(grpc.WithInsecure())
//	require.NoError(t, err)
//	defer client.Stop()
//
//	bridgeAddres, err := common.Base58ToAddress(bridgeHex)
//	require.NoError(t, err)
//
//	bridge := NewBridgeContract(client, bridgeAddres)
//	resource, err := hexutil.Decode(resourceHex)
//	require.NoError(t, err)
//
//	actualAddress, err := bridge.GetHandlerAddressForResourceID(toFixed(resource))
//	require.NoError(t, err)
//
//	expectedAddress, err := common.Base58ToAddress(expectedHandlerBase58)
//	require.NoError(t, err)
//
//	require.Equal(t, expectedAddress.String(), actualAddress.String())
//}
//
//func toFixed(in []byte) [32]byte {
//	res := (*[32]byte)(in)
//	return *res
//}
