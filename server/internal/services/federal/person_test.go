package federal

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/services/net"
	"github.com/kidommoc/gustrody/internal/test"
	"github.com/kidommoc/gustrody/internal/utils"
)

var psncfg = config.Config{
	Scheme: "http",
	Domain: "127.0.0.1:7999",
}

func site() string {
	return psncfg.Scheme + "://" + psncfg.Domain
}

type outerPersonInput struct {
	Raw  string
	Ud   models.UD
	Want PersonObj
}

var outerPersonTable = map[string]outerPersonInput{
	site() + "/users/polarbear": {
		Ud: models.UD{Username: "polarbear", Domain: psncfg.Domain},
		Raw: `{
			    "@context": ["https://www.w3.org/ns/activitystreams","https://w3id.org/security/v1",{"manuallyApprovesFollowers": "as:manuallyApprovesFollowers","toot": "http://joinmastodon.org/ns#","featured": {"@id": "toot:featured","@type": "@id"},"featuredTags": {"@id": "toot:featuredTags","@type": "@id"},"alsoKnownAs": {"@id": "as:alsoKnownAs","@type": "@id"},"movedTo": {"@id": "as:movedTo","@type": "@id"},"schema": "http://schema.org#","PropertyValue": "schema:PropertyValue","value": "schema:value","discoverable": "toot:discoverable","Device": "toot:Device","Ed25519Signature": "toot:Ed25519Signature","Ed25519Key": "toot:Ed25519Key","Curve25519Key": "toot:Curve25519Key","EncryptedMessage": "toot:EncryptedMessage","publicKeyBase64": "toot:publicKeyBase64","deviceId": "toot:deviceId","claim": {"@type": "@id","@id": "toot:claim"},"fingerprintKey": {"@type": "@id","@id": "toot:fingerprintKey"},"identityKey": {"@type": "@id","@id": "toot:identityKey"},"devices": {"@type": "@id","@id": "toot:devices"},"messageFranking": "toot:messageFranking","messageType": "toot:messageType","cipherText": "toot:cipherText","suspended": "toot:suspended","memorial": "toot:memorial","indexable": "toot:indexable","focalPoint": {"@container": "@list","@id": "toot:focalPoint"}}],
			    "id": "https://cookie.eat/users/polarbear", "type": "Person",
			    "following": "https://cookie.eat/users/polarbear/following", "followers": "https://cookie.eat/users/polarbear/followers",
			    "inbox": "` + site() + `/inbox", "outbox": "https://cookie.eat/users/polarbear/outbox",
			    "featured": "https://cookie.eat/users/polarbear/collections/featured", "featuredTags": "https://cookie.eat/users/polarbear/collections/tags",
			    "preferredUsername": "polarbear", "name": "BearP",
			    "summary": "<p>Live ~ in peace</p>",
			    "url": "https://cookie.eat/@polarbear",
			    "manuallyApprovesFollowers": true, "discoverable": false, "indexable": false,
			    "published": "2022-11-26T00:00:00Z", "memorial": false,
			    "devices": "https://cookie.eat/users/polarbear/collections/devices",
			    "alsoKnownAs": ["https://m.cmx.im/users/polarbear"],
			    "publicKey": {
			      "id": "https://cookie.eat/users/polarbear#main-key",
			      "owner": "https://cookie.eat/users/polarbear",
			      "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArbSOeaSC+Yl8G0RCnTNo\nqameZyuATYBg7Sp9N1II/lwjIVZZlpTvo74n/NASdB+MXUME+jCNick/emX0uvj0\nSv8eK26I7NDry3Zr5SCzvS7C7qxG1o7usK8AlwEcAOZCRLAA9x07Z5+8eXV4cOi8\nODtOeee7YgL9PSDxA1ieznLNIKCZFIBYCKEUnbuL/ip1JJPAFDVrb39M7qsHh4B/\nqbEmF7DadoQIsSEteZyEe2GVWNs2t4XWLOt18THRgCJQ/LRVizZ1Fo4slhLffI/f\nYwa9uqj2Rhf+a1NJV0Ytu6BrnB7x+jQ+82Ch/7IyOr50rQiPMW5o2K/lzx/v73T0\nlQIDAQAB\n-----END PUBLIC KEY-----\n"
			    },
			    "tag": [], "attachment": [],
			    "icon": {
			      "type": "Image", "mediaType": "image/jpg",
			      "url": "` + site() + `/img/4a855f1c2a95b0ec.jpg"
			    },
			    "image": {
			      "type": "Image", "mediaType": "image/jpeg",
			      "url": "https://img.cookie.eat/accounts/headers/109/409/091/728/348/948/original/f1837420bc37f066.png"
			    }
			  }`,
		Want: PersonObj{
			Object: Object{
				Type: "Person", ID: "https://cookie.eat/users/polarbear",
			},
			Username: "polarbear", Nickname: "BearP",
			Summary: "<p>Live ~ in peace</p>",
			Inbox:   site() + "/inbox", Outbox: "https://cookie.eat/users/polarbear/outbox",
			Following: "https://cookie.eat/users/polarbear/following", Followers: "https://cookie.eat/users/polarbear/followers",
			Icon: MediaObj{
				Object:    Object{Type: "Image"},
				MediaType: "image/jpeg",
				Url:       site() + "/img/2f29f3d7572f9d744772043dd703fd9fca289c673305a3435c532a25bc02df61.jpeg",
			},
			PublicKey: PublicKeyObj{
				Object: Object{ID: "https://cookie.eat/users/polarbear#main-key"},
				Owner:  "https://cookie.eat/users/polarbear",
				Pem:    "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArbSOeaSC+Yl8G0RCnTNo\nqameZyuATYBg7Sp9N1II/lwjIVZZlpTvo74n/NASdB+MXUME+jCNick/emX0uvj0\nSv8eK26I7NDry3Zr5SCzvS7C7qxG1o7usK8AlwEcAOZCRLAA9x07Z5+8eXV4cOi8\nODtOeee7YgL9PSDxA1ieznLNIKCZFIBYCKEUnbuL/ip1JJPAFDVrb39M7qsHh4B/\nqbEmF7DadoQIsSEteZyEe2GVWNs2t4XWLOt18THRgCJQ/LRVizZ1Fo4slhLffI/f\nYwa9uqj2Rhf+a1NJV0Ytu6BrnB7x+jQ+82Ch/7IyOr50rQiPMW5o2K/lzx/v73T0\nlQIDAQAB\n-----END PUBLIC KEY-----\n",
			},
		},
	},
	site() + "/users/9b45fhe9zi": {
		Ud: models.UD{Username: "9621", Domain: psncfg.Domain},
		Raw: `{
			    "@context": ["https://www.w3.org/ns/activitystreams","https://w3id.org/security/v1",{"Key": "sec:Key","manuallyApprovesFollowers": "as:manuallyApprovesFollowers","sensitive": "as:sensitive","Hashtag": "as:Hashtag","quoteUrl": "as:quoteUrl","toot": "http://joinmastodon.org/ns#","Emoji": "toot:Emoji","featured": "toot:featured","discoverable": "toot:discoverable","schema": "http://schema.org#","PropertyValue": "schema:PropertyValue","value": "schema:value","mickey": "https://mickey-hub.net/ns#","_mickey_content": "mickey:_mickey_content","_mickey_quote": "mickey:_mickey_quote","_mickey_reaction": "mickey:_mickey_reaction","_mickey_votes": "mickey:_mickey_votes","_mickey_summary": "mickey:_mickey_summary","isCat": "mickey:isCat","vcard": "http://www.w3.org/2006/vcard/ns#"}],
			    "type": "Person", "id": "https://mickey.io/users/9b45fhe9zi",
			    "inbox": "https://mickey.io/users/9b45fhe9zi/inbox", "outbox": "https://mickey.io/users/9b45fhe9zi/outbox",
			    "followers": "https://mickey.io/users/9b45fhe9zi/followers", "following": "https://mickey.io/users/9b45fhe9zi/following",
			    "featured": "https://mickey.io/users/9b45fhe9zi/collections/featured",
			    "sharedInbox": "https://mickey.io/inbox",
			    "endpoints": { "sharedInbox": "` + site() + `/inbox" },
			    "url": "https://mickey.io/@9621",
			    "preferredUsername": "9621", "name": "9621ちゃん.io:blobcatmeataww:",
			    "summary": "<p><div><span>いんたーねっとでおえかきするてんし<br></span>​:ablobcatreachflip:​と​:murakamisan_shock:​<span>がすき/成人済<br></span>​:blobcatmlem:​の絵<a href=\"https://mickey.io/clips/9e8ghap0vu\">https://mickey.io/clips/9e8ghap0vu</a><span><br></span>​:icon_murakamisan:​の絵<a href=\"https://mickey.io/clips/9c020udrhl\">https://mickey.io/clips/9c020udrhl</a><span><br>♡┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈♡<br>オタクしておやつ食べる:</span><a href=\"https://mickey.io/@9621@nijimiss.moe\" class=\"u-url mention\">@9621@nijimiss.moe</a><span><br>絵を描いてる方:</span><a href=\"https://mickey.io/@9621@mickey.art\" class=\"u-url mention\">@9621@mickey.art</a><span><br></span><a href=\"https://mickey.io/@9621/pages/1688831646136\">他鯖のアカウント一覧</a><span><br></span>​:sitteokou:​ 9621はくろつちと読みます</div></p>",
			    "_mickey_summary": "<center>いんたーねっとでおえかきするてんし\n:ablobcatreachflip:と:murakamisan_shock:がすき/成人済\n:blobcatmlem:の絵https://mickey.io/clips/9e8ghap0vu\n:icon_murakamisan:の絵https://mickey.io/clips/9c020udrhl\n♡┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈♡\nオタクしておやつ食べる:@9621@nijimiss.moe\n絵を描いてる方:@9621@mickey.art\n[他鯖のアカウント一覧](https://mickey.io/@9621/pages/1688831646136)\n:sitteokou: 9621はくろつちと読みます</center>",
			    "icon": {
			      "type": "Image",
			      "url": "` + site() + `/img/1b4d4468740344d5.jpg",
			      "sensitive": false, "name": null
			    },
			    "image": {
			      "type": "Image",
			      "url": "https://media.mickeyusercontent.com/mickey/e9f1e104-32de-442f-a8d0-7e0ab1c88ca0.png",
			      "sensitive": false, "name": null
			    },
			    "tag": [ { "id": "https://mickey.io/emojis/blobcatmeataww", "type": "Emoji", "name": ":blobcatmeataww:", "updated": "2023-10-30T00:31:18.921Z", "icon": { "type": "Image", "mediaType": "image/png", "url": "https://media.mickeyusercontent.com/emoji/blobcatmeataww.png" } }, { "id": "https://mickey.io/emojis/ablobcatreachflip", "type": "Emoji", "name": ":ablobcatreachflip:", "updated": "2023-10-30T00:33:33.205Z", "icon": { "type": "Image", "mediaType": "image/apng", "url": "https://media.mickeyusercontent.com/emoji/ablobcatreachflip.apng" } }, { "id": "https://mickey.io/emojis/murakamisan_shock", "type": "Emoji", "name": ":murakamisan_shock:", "updated": "2023-10-29T23:32:32.494Z", "icon": { "type": "Image", "mediaType": "image/gif", "url": "https://media.mickeyusercontent.com/mickey/16bd73b2-5a15-44c5-84fb-90cba740ed33.gif" } }, { "id": "https://mickey.io/emojis/blobcatmlem", "type": "Emoji", "name": ":blobcatmlem:", "updated": "2023-10-30T00:31:18.921Z", "icon": { "type": "Image", "mediaType": "image/png", "url": "https://media.mickeyusercontent.com/emoji/blobcatmlem.png" } }, { "id": "https://mickey.io/emojis/icon_murakamisan", "type": "Emoji", "name": ":icon_murakamisan:", "updated": "2023-10-29T23:32:32.494Z", "icon": { "type": "Image", "mediaType": "image/png", "url": "https://media.mickeyusercontent.com/emoji/icon_murakamisan.png" } }, { "id": "https://mickey.io/emojis/sitteokou", "type": "Emoji", "name": ":sitteokou:", "updated": "2023-04-28T19:34:46.259Z", "icon": { "type": "Image", "mediaType": "image/jpeg", "url": "https://media.mickeyusercontent.com/mickey/webpublic-e80bab51-0c91-4726-9964-53a80cad1901.jpg" } }, { "id": "https://mickey.io/emojis/unicode_1d54f_bg_black", "type": "Emoji", "name": ":unicode_1d54f_bg_black:", "updated": "2023-07-25T06:52:49.597Z", "icon": { "type": "Image", "mediaType": "image/png", "url": "https://media.mickeyusercontent.com/io/5b135545-e11f-45e1-bcdf-66b39adb57e7.png" } }, { "id": "https://mickey.io/emojis/booth", "type": "Emoji", "name": ":booth:", "updated": "2023-02-11T21:26:09.008Z", "icon": { "type": "Image", "mediaType": "image/png", "url": "https://media.mickeyusercontent.com/emoji/booth.png" } } ],
			    "manuallyApprovesFollowers": false, "discoverable": true,
			    "publicKey": {
			      "id": "https://mickey.io/users/9b45fhe9zi#main-key",
			      "type": "Key",
			      "owner": "https://mickey.io/users/9b45fhe9zi",
			      "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtI/hmamisKQkmSYQKf+M\n43uUh5e25G6kBAVwbnh6/qeMzG0U4Ws7+yqEN1uL8i+rKmgdPbIWKR8WMQT++nZJ\nz/+LJyBx2X8g58FaWKu6vrp589ijGhtofrO6gkD0EYh0JESVTIBiLogLsnXlRsn+\nJT/qA6rvqXb0JgYvSbqvgQyouo/pmyA2II7sonQiyOX1ptdbiKBNBkWiTDGKauFR\nRCic1To5bvNzbKlMSMOkqbY+CJwgZtog8kDF70SDDgyZiS8Xxn2Nvt1irzzyCYCF\nznpRueE/VV+MEIOoZWl5Hm+KicJ43PBhGo0GaIF3SNUplOueTfYcBfV1Jn+1ir9S\nlZQ7x7C31AwJDV3xqQ0/RyfxVCRoJp2py5jmzMColFMsIvDqm46B7W+a3Shfcxad\nsh4F3rpweXaE+f9zHfT1ZJGgSUCLzL3drXGFj1GDS5Xuw9svKviSo8TU5kHI59BK\nFUY12yy07wDRqO1mHKwNktzhkRFa0JOHQbrQbDVhPmHDsgHnZ6E3RJ8dRmnfRu4c\n5wQ8pzNxUTzov+VfuZkVlpSOKOZUc2qnGMT+azWCHi/a/QHPomu0F0HJHtiw2nwh\n0sp9gVYPfugl77sWo/VhWGLKXpTGwt+oFOvvBBiveB4NrxV4Xw9pYpbL2zfIPMP7\nsy5nnmJVsi9G8QcYpBFC/RcCAwEAAQ==\n-----END PUBLIC KEY-----\n"
			    },
			    "isCat": false,
			    "attachment": [ { "type": "PropertyValue", "name": ":unicode_1d54f_bg_black:", "value": "<a href=\"https://twitter.com/9621chan\" rel=\"me nofollow noopener\" target=\"_blank\">https://twitter.com/9621chan</a>" }, { "type": "PropertyValue", "name": ":booth:", "value": "<a href=\"https://9621.booth.pm/\" rel=\"me nofollow noopener\" target=\"_blank\">https://9621.booth.pm/</a>" }, { "type": "PropertyValue", "name": "Suzuri", "value": "<a href=\"https://suzuri.jp/9621/home\" rel=\"me nofollow noopener\" target=\"_blank\">https://suzuri.jp/9621/home</a>" }, { "type": "PropertyValue", "name": "Bluesky", "value": "<a href=\"https://bsky.app/profile/9621.bsky.social\" rel=\"me nofollow noopener\" target=\"_blank\">https://bsky.app/profile/9621.bsky.social</a>" } ],
			    "vcard:bday": "9621-05-22"
			  }`,
		Want: PersonObj{
			Object: Object{
				Type: "Person", ID: "https://mickey.io/users/9b45fhe9zi",
			},
			Username: "9621", Nickname: "9621ちゃん.io:blobcatmeataww:",
			Summary: "<p><div><span>いんたーねっとでおえかきするてんし<br></span>​:ablobcatreachflip:​と​:murakamisan_shock:​<span>がすき/成人済<br></span>​:blobcatmlem:​の絵<a href=\"https://mickey.io/clips/9e8ghap0vu\">https://mickey.io/clips/9e8ghap0vu</a><span><br></span>​:icon_murakamisan:​の絵<a href=\"https://mickey.io/clips/9c020udrhl\">https://mickey.io/clips/9c020udrhl</a><span><br>♡┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈♡<br>オタクしておやつ食べる:</span><a href=\"https://mickey.io/@9621@nijimiss.moe\" class=\"u-url mention\">@9621@nijimiss.moe</a><span><br>絵を描いてる方:</span><a href=\"https://mickey.io/@9621@mickey.art\" class=\"u-url mention\">@9621@mickey.art</a><span><br></span><a href=\"https://mickey.io/@9621/pages/1688831646136\">他鯖のアカウント一覧</a><span><br></span>​:sitteokou:​ 9621はくろつちと読みます</div></p>",
			Inbox:   "https://mickey.io/users/9b45fhe9zi/inbox", Outbox: "https://mickey.io/users/9b45fhe9zi/outbox",
			Endpoints: struct {
				SharedInbox string "json:\"sharedInbox\""
			}{
				SharedInbox: site() + "/inbox",
			},
			Following: "https://mickey.io/users/9b45fhe9zi/following", Followers: "https://mickey.io/users/9b45fhe9zi/followers",
			Icon: MediaObj{
				Object:    Object{Type: "Image"},
				MediaType: "image/jpeg",
				Url:       site() + "/img/ac7659cd1bfd8337f6848fe2de04e02ee22749e7ca797e61ab23e8bd979e97a8.jpeg",
			},
			PublicKey: PublicKeyObj{
				Object: Object{ID: "https://mickey.io/users/9b45fhe9zi#main-key"},
				Owner:  "https://mickey.io/users/9b45fhe9zi",
				Pem:    "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtI/hmamisKQkmSYQKf+M\n43uUh5e25G6kBAVwbnh6/qeMzG0U4Ws7+yqEN1uL8i+rKmgdPbIWKR8WMQT++nZJ\nz/+LJyBx2X8g58FaWKu6vrp589ijGhtofrO6gkD0EYh0JESVTIBiLogLsnXlRsn+\nJT/qA6rvqXb0JgYvSbqvgQyouo/pmyA2II7sonQiyOX1ptdbiKBNBkWiTDGKauFR\nRCic1To5bvNzbKlMSMOkqbY+CJwgZtog8kDF70SDDgyZiS8Xxn2Nvt1irzzyCYCF\nznpRueE/VV+MEIOoZWl5Hm+KicJ43PBhGo0GaIF3SNUplOueTfYcBfV1Jn+1ir9S\nlZQ7x7C31AwJDV3xqQ0/RyfxVCRoJp2py5jmzMColFMsIvDqm46B7W+a3Shfcxad\nsh4F3rpweXaE+f9zHfT1ZJGgSUCLzL3drXGFj1GDS5Xuw9svKviSo8TU5kHI59BK\nFUY12yy07wDRqO1mHKwNktzhkRFa0JOHQbrQbDVhPmHDsgHnZ6E3RJ8dRmnfRu4c\n5wQ8pzNxUTzov+VfuZkVlpSOKOZUc2qnGMT+azWCHi/a/QHPomu0F0HJHtiw2nwh\n0sp9gVYPfugl77sWo/VhWGLKXpTGwt+oFOvvBBiveB4NrxV4Xw9pYpbL2zfIPMP7\nsy5nnmJVsi9G8QcYpBFC/RcCAwEAAQ==\n-----END PUBLIC KEY-----\n",
			},
		},
	},
	site() + "/users/snowwhite": {
		Ud: models.UD{Username: "snowwhite", Domain: psncfg.Domain},
		Raw: `{
			    "@context": [ "https://www.w3.org/ns/activitystreams", "https://w3id.org/security/v1", { "manuallyApprovesFollowers": "as:manuallyApprovesFollowers", "toot": "http://joinmastodon.org/ns#", "featured": { "@id": "toot:featured", "@type": "@id" }, "featuredTags": { "@id": "toot:featuredTags", "@type": "@id" }, "alsoKnownAs": { "@id": "as:alsoKnownAs", "@type": "@id" }, "movedTo": { "@id": "as:movedTo", "@type": "@id" }, "schema": "http://schema.org#", "PropertyValue": "schema:PropertyValue", "value": "schema:value", "discoverable": "toot:discoverable", "Device": "toot:Device", "Ed25519Signature": "toot:Ed25519Signature", "Ed25519Key": "toot:Ed25519Key", "Curve25519Key": "toot:Curve25519Key", "EncryptedMessage": "toot:EncryptedMessage", "publicKeyBase64": "toot:publicKeyBase64", "deviceId": "toot:deviceId", "claim": { "@type": "@id", "@id": "toot:claim" }, "fingerprintKey": { "@type": "@id", "@id": "toot:fingerprintKey" }, "identityKey": { "@type": "@id", "@id": "toot:identityKey" }, "devices": { "@type": "@id", "@id": "toot:devices" }, "messageFranking": "toot:messageFranking", "messageType": "toot:messageType", "cipherText": "toot:cipherText", "suspended": "toot:suspended", "memorial": "toot:memorial", "indexable": "toot:indexable" } ],
			    "id": "https://exam.ple/users/snowwhite", "type": "Person",
			    "following": "https://exam.ple/users/snowwhite/following", "followers": "https://exam.ple/users/snowwhite/followers",
			    "inbox": "https://exam.ple/users/snowwhite/inbox", "outbox": "https://exam.ple/users/snowwhite/outbox",
			    "featured": "https://exam.ple/users/snowwhite/collections/featured", "featuredTags": "https://exam.ple/users/snowwhite/collections/tags",
			    "preferredUsername": "snowwhite", "name": "snowwhite",
			    "summary": "",
			    "url": "https://exam.ple/@snowwhite",
			    "manuallyApprovesFollowers": true, "discoverable": false, "indexable": false,
			    "published": "2021-10-29T00:00:00Z", "memorial": false,
			    "devices": "https://exam.ple/users/snowwhite/collections/devices",
			    "publicKey": {
			      "id": "https://exam.ple/users/snowwhite#main-key",
			      "owner": "https://exam.ple/users/snowwhite",
			      "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvgtlqWW35ANKYh7CgH2a\n+/nsxUisx/a5w29PnP/QhNfEasH3Bur1FY0zFER9Prv7th6eYT9/fNLut3a8ICU0\nSRyX1eBW7t2Y252zEhj+Fcgvpf0SieUza3kwC+8EftSQWxs28IuP14IIHzwUaKnG\nTS0am5n9i3dIbJhvynV5ekeqGlf8xqG25ser+ZQwPwQEOiLir5LqTyOh4Ve3BP/Z\ne5yxkvFfShmtE5PKenxSXoKb1csYYQAGbl1iFq9GkXmPTUIf0Oluxsr/xj12TYFb\neE1X0H8UNYAHWc+S7CukVp+guL1PTu+Ln5u7SzolQLzJHgBOAWy5EjqBBNDcutgH\nZwIDAQAB\n-----END PUBLIC KEY-----\n"
			    },
			    "tag": [],
			    "attachment": [],
			    "endpoints": { "sharedInbox": "` + site() + `/inbox" }
			  }`,
		Want: PersonObj{
			Object: Object{
				Type: "Person", ID: "https://exam.ple/users/snowwhite",
			},
			Username: "snowwhite", Nickname: "snowwhite",
			Summary: "",
			Inbox:   "https://exam.ple/users/snowwhite/inbox", Outbox: "https://exam.ple/users/snowwhite/outbox",
			Endpoints: struct {
				SharedInbox string "json:\"sharedInbox\""
			}{
				SharedInbox: site() + "/inbox",
			},
			Following: "https://exam.ple/users/snowwhite/following", Followers: "https://exam.ple/users/snowwhite/followers",
			Icon: MediaObj{},
			PublicKey: PublicKeyObj{
				Object: Object{ID: "https://exam.ple/users/snowwhite#main-key"},
				Owner:  "https://exam.ple/users/snowwhite",
				Pem:    "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvgtlqWW35ANKYh7CgH2a\n+/nsxUisx/a5w29PnP/QhNfEasH3Bur1FY0zFER9Prv7th6eYT9/fNLut3a8ICU0\nSRyX1eBW7t2Y252zEhj+Fcgvpf0SieUza3kwC+8EftSQWxs28IuP14IIHzwUaKnG\nTS0am5n9i3dIbJhvynV5ekeqGlf8xqG25ser+ZQwPwQEOiLir5LqTyOh4Ve3BP/Z\ne5yxkvFfShmtE5PKenxSXoKb1csYYQAGbl1iFq9GkXmPTUIf0Oluxsr/xj12TYFb\neE1X0H8UNYAHWc+S7CukVp+guL1PTu+Ln5u7SzolQLzJHgBOAWy5EjqBBNDcutgH\nZwIDAQAB\n-----END PUBLIC KEY-----\n",
			},
		},
	},
}

func startPersonServer(ch chan string) {
	app := fiber.New()

	app.Static("/img", "./testfile")

	app.Use("/users/:username", func(c *fiber.Ctx) error {
		id := site() + c.Path()
		if outerPersonTable[id].Raw != "" {
			c.Set("Content-Type", "application/activity+json")
			return c.Send([]byte(outerPersonTable[id].Raw))
		}
		return c.SendStatus(fiber.StatusNotFound)
	})

	app.Post("/inbox", func(c *fiber.Ctx) error {
		ch <- string(c.BodyRaw())
		return c.SendStatus(fiber.StatusOK)
	})

	app.Listen(psncfg.Domain)
}

func TestGetForeignPerson(t *testing.T) {
	logger := test.NewMockingLogger(t)
	mUFgnDb := newMockingUserForeignDb()
	dbs := FederalDbs{
		UserForeign: mUFgnDb,
	}
	netService := net.NewNetService(logger)
	fileService := newMockingFileService(psncfg)
	service := NewService(netService, fileService, dbs, psncfg, logger)

	ch := make(chan string, 1)
	go startPersonServer(ch)
	time.Sleep(2 * time.Second) // wait for mocking foreign server starting
	t.Log("...starts.")

	t.Run("Test outer person", func(t *testing.T) {
		for k, v := range outerPersonTable {
			got, err := service.GetForeignPerson(k)
			test.AssertNoError(t, err)
			got.Context = nil
			got.PublicKey.Context = nil
			got.PublicKey.Type = ""
			test.AssertEqual(t, v.Want, got)
		}
	})

	t.Run("Test ud id converting", func(t *testing.T) {
		for _, v := range outerPersonTable {
			gotUd := service.idToUD(v.Want.ID)
			test.AssertEqual(t, v.Ud, gotUd)
			gotId := service.udToID(v.Ud)
			test.AssertEqual(t, v.Want.ID, gotId)
		}
	})
}

func TestForeignFollow(t *testing.T) {
	logger := test.NewMockingLogger(t)
	foSig := make(chan bool, 1)
	mUAccDb := newMockingUserAccountDb()
	mUFlwDb := newMockingUserFollowDb(foSig)
	mUFgnDb := newMockingUserForeignDb()
	dbs := FederalDbs{
		UserAccount: mUAccDb, UserFollow: mUFlwDb,
		UserForeign: mUFgnDb,
	}
	netService := net.NewNetService(logger)
	fileService := newMockingFileService(psncfg)
	service := NewService(netService, fileService, dbs, psncfg, logger)

	ch := make(chan string, 1)
	go startPersonServer(ch)
	time.Sleep(2 * time.Second) // wait for mocking foreign server starting
	t.Log("...starts.")

	// init local account
	pub, pri := utils.NewKeyPair()
	mUAccDb.data["u1"] = struct {
		pub string
		pri string
	}{pub, pri}
	// init foreign account
	for k := range outerPersonTable {
		service.GetForeignPerson(k)
	}

	t.Run("Test request follow", func(t *testing.T) {
		template := `{"@context":["https://www.w3.org/ns/activitystreams","https://w3id.org/security/v1"],"id":"http://127.0.0.1:7999/users/u1#follow/%s","type":"Follow","actor":"` + site() + `/users/u1","object":"%s"}`
		for _, v := range outerPersonTable {
			err := service.RequestFollow(models.NewUD("u1"), v.Ud)
			test.AssertNoError(t, err)
			got := <-ch
			test.AssertEqual(t, fmt.Sprintf(template, v.Ud, v.Want.ID), got)

			req := Activity{
				Object: Object{},
				Actor:  v.Want.ID,
				Target: site() + "/users/theuser1",
			}
			actorUD := service.idToUD(v.Want.ID)
			targetUD := models.NewUD("theuser1")
			actor, target, err := service.ParseFollowRequest(&req)
			test.AssertNoError(t, err)
			test.AssertEqual(t, actorUD, actor)
			test.AssertEqual(t, targetUD, target)
		}
	})

	tmpIdReg := regexp.MustCompile(`.+"id":"(.+?)".*`)

	t.Run("Test undo follow request", func(t *testing.T) {
		template := `{"@context":["https://www.w3.org/ns/activitystreams","https://w3id.org/security/v1"],"id":"%s","type":"Undo","actor":"` + site() + `/users/u1","object":"` + site() + `/users/u1#follow/%s"}`
		for _, v := range outerPersonTable {
			err := service.UndoFollow(models.NewUD("u1"), v.Ud)
			test.AssertNoError(t, err)
			done := <-foSig
			test.AssertEqual(t, true, done)
			got := <-ch
			match := tmpIdReg.FindStringSubmatch(got)
			test.AssertEqual(t, 2, len(match))
			test.AssertEqual(t, fmt.Sprintf(template, match[1], v.Ud), got)
		}
	})

	t.Run("Test accept follow request", func(t *testing.T) {
		template := `{"@context":["https://www.w3.org/ns/activitystreams","https://w3id.org/security/v1"],"id":"%s","type":"Accept","actor":"` + site() + `/users/u1","object":"tempID"}`
		for _, v := range outerPersonTable {
			u1 := models.NewUD("u1")
			err := service.AcceptFollow(u1, v.Ud, "tempID")
			test.AssertNoError(t, err)
			done := <-foSig
			test.AssertEqual(t, true, done)
			got := <-ch
			match := tmpIdReg.FindStringSubmatch(got)
			test.AssertEqual(t, 2, len(match))
			test.AssertEqual(t, fmt.Sprintf(template, match[1]), got)
		}
	})

	t.Run("Test reject follow request", func(t *testing.T) {
		template := `{"@context":["https://www.w3.org/ns/activitystreams","https://w3id.org/security/v1"],"id":"%s","type":"Reject","actor":"` + site() + `/users/u1","object":"tempID"}`
		for _, v := range outerPersonTable {
			u1 := models.NewUD("u1")
			err := service.RejectFollow(u1, v.Ud, "tempID")
			test.AssertNoError(t, err)
			select {
			case done := <-foSig:
				if done == true {
					t.Errorf("any follow action is done, which is not expected.")
				}
			default:
			}
			got := <-ch
			match := tmpIdReg.FindStringSubmatch(got)
			test.AssertEqual(t, 2, len(match))
			test.AssertEqual(t, fmt.Sprintf(template, match[1]), got)
		}
	})
}
