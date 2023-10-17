// Copyright 2021 The sdvn Authors
// This file is part of the sdvn library.
//
// The sdvn library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The sdvn library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the sdvn library. If not, see <http://www.gnu.org/licenses/>.

package params

import "github.com/seaskycheng/sdvn/common"

// MainnetBootnodes are the enode URLs of the P2P bootstrap nodes running on
// the main sdvn network.
var MainnetBootnodes = []string{
	// sdvn Foundation Go Bootnodes
	"enode://76c6654de53a581fecdee60b72371a8c6eb54a77793eb6627d9515222a230fa27408f2f65f23022e4fe050d100ed4f932d1c4b95bb0a9203723959ecdf896efb@44.238.157.20:30311",
	"enode://096e8baabd2ca1cc6b4c5963314131eb998a171884ee07ea2ec5858536272a69a5c558ba5dfdac03511b9a1298282aaadfcc9afc1e7d5f88637d134ea976ef4b@18.198.18.94:30311",
	"enode://4496e5649cd64ad38c264699599fafcb0a6c377f7dd7aadfd2c028efebf78585aca8829e667d2bb6290da6aeff00bd88aa5b2794e3dc2b58badbfc10c4de7826@18.166.186.10:30311",
	"enode://c15a807771cc22d1e864087ada3c1e07782690636ed9979d1648bbf9c4954b78af4e6e753370a08898e942e9d68ba5a773e542b4adfe855179ce1bfc653a2f0c@65.19.174.213:30311",
	"enode://f94e7126d29676e7c795c3f5bf067931709cf12cdc91990918a7c030cc81a0377c0b0c950a6f9ea00e0990cfc8e394b3128c49882a4848d969e3f2feb1f42bc3@65.19.174.213:30312",
	"enode://88b918d54af1c9a45e3306a75275bae153b568cf2c23b37d67fd09613d4921229a3433f86d366966ee42339b0a5d4025e0aecaa19600ca63a845b512d04de5d5@65.19.174.250:30311",
	"enode://11c9ca59a38641742eef8833875af531de200c4362a94b274a0ae4746203a581374c8f0d21b0d53bc39d4f4619ead83a331368d831bfd603c0743d90a134669e@65.19.174.250:30312",
}

// TestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Testnet test network.
var TestnetBootnodes = []string{
	"enode://2ffed1bb6b475259c1448dc93b639569886999e51ade144451877a706d2a9b71eff8eb067d289fde48ba4807370034d851553746fac8816af27f5a922703e2e4@127.0.0.1:30311",
}

var V5Bootnodes = []string{
}

const dnsPrefix = "enrtree://AKA3AM6LPBYEUDMVNU3BSVQJ5AD45Y7YPOHJLEF6W26QOE4VTUDPE@"

// KnownDNSNetwork returns the address of a public DNS-based node list for the given
// genesis hash and protocol. See https://github.com/ethereum/discv4-dns-lists for more
// information.
func KnownDNSNetwork(genesis common.Hash, protocol string) string {
	var net string
	switch genesis {
	case MainnetGenesisHash:
		net = "mainnet"
	case TestnetGenesisHash:
		net = "testnet"
	default:
		return ""
	}
	return dnsPrefix + protocol + "." + net + ".ethdisco.net"
}
