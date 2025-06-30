package service

import (
	"djj-inventory-system/assets"
	"encoding/base64"
)

// 在需要的时候这样拿
var LogoBase64 = base64.StdEncoding.EncodeToString(assets.LogoPNG)
