package m3u8

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	STREAMMASTER = `#EXTM3U
#EXT-X-TWITCH-INFO:NODE="video-edge-d54460.ams02",MANIFEST-NODE-TYPE="weaver_cluster",MANIFEST-NODE="video-weaver.vie02",SUPPRESS="true",SERVER-TIME="1721130391.23",TRANSCODESTACK="2017TranscodeX264_V2",TRANSCODEMODE="cbr_v1",USER-IP="178.149.124.105",SERVING-ID="2ad06086caf343b99b648d8951d3a1e8",CLUSTER="ams02",ABS="true",VIDEO-SESSION-ID="1645180638639471244",BROADCAST-ID="44535804459",STREAM-TIME="16967.234778",B="false",USER-COUNTRY="RS",MANIFEST-CLUSTER="vie02",ORIGIN="pdx05",C="aHR0cHM6Ly92aWRlby1lZGdlLWRjODQwOC5wZHgwMS5hYnMuaGxzLnR0dm53Lm5ldC92MS9zZWdtZW50L0NnMWlhcmdyd0lBMXdlMkc3dDNEeXpkUzNwandjRkJQSFNyUFpkMG13ek1yVzhienZURmZhNDFFMVFFU2JIOUpOU3ZSYmpZUzdkSTNUTEV1ZGtYdHRneWRWRUZ0MXlNWHRQSEpiSUxBYy1zRE5qTWhNNEY5RlI3NXVjdGNPb1NxNi1JUG43a3RKbjdxVGFySFFXOFZJeFdZTlFmaWpCSS1FLUhLbm90OFpuU3VVNmRNTXRFOEhsdjdBSVpDVXNQbVBtOWEwNV9PZlJDWHZYVFdja0JubmtKaE11UV82dUNjRWNsWmdycUp6M2tORU1GSFB0VGM4ZDVnQUxBOEhLbWdHaTFLVGtjeUJKdmoxbDRkdjhvZmxZZnFFVWF2c3pVeUtLeFhVamQzUkgtU1FTS1BwZ29DUjY1TlJpUXlNU0VKa19hb0R6T1pRb1dGZ1J6M2NxUzBMOTNPbzBMS0Y1TFlMRVBEdGNDLUpaNmNMRHRaWmR2Q1JWLWt5STN1b2pOa0dwNDlJMjg5OXd2QUxySW9JVjREYllKWHB1VmMtWTRKaU1UUWw3VVZMNS1rT1dtUG5EM1MxT2FwSGdxamJoaHhVdkNSZ25RbDEyMUtPOVhsU3RFSUVDZGJvSUpBVXVSdGV1MGlJOHhTZ1VrVEZwUHhIWE9hWHBVZGNBRUY0UkMyYU1jM0djdlBkUWVNRWUtam56OXVSbVNRbnFMVjZSTWtRWmlWbVU3ODNlY2doZi1ZN1pqYUxuN1FlUVFrZnQyYlVDU05JYTFWTHB6SXRRbzV3Z193WXZzTkJuTmNad1JlZGZibUwzOFIyMXh1RkdaUFpLTDdsQW5aX3IzaWdRQnh6dEFPeGlZa1Bac2VCT25yRkpiT25UTS1JUU1Uaml6b0tUaWZZZFFnLXEwYVR1a05oTlBpR1BnQUhVU2FyZGR5OWNDZ0MyWmdFMnRtazdadDZ3XzR5bkhrbEt5NUxMLWN2bXJZeFd2X3d4Y0Z3TG83N19hbW53dkk1aGYzajc3X1J6STdWWUhBMG5jdGxORy50cw",D="false"
#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID="chunked",NAME="1080p60 (source)",AUTOSELECT=YES,DEFAULT=YES
#EXT-X-STREAM-INF:BANDWIDTH=7021569,RESOLUTION=1920x1080,CODECS="avc1.64002A,mp4a.40.2",VIDEO="chunked",FRAME-RATE=60.000
https://video-weaver.vie02.hls.ttvnw.net/v1/playlist/CvEEeTMRRPmeg6Ieg2x5qD5Sdsj_Pb9nlNC5GOdgbR94WF3P15ekN2KXRirvrQeHd0QxXjqeGCdtqH56Fs0kwJDg2YN7mJcaPSnRRa2icGWOAbr_yleYlkO6X95hmqIKN0EN7eYCztOc2na35SGTfX406d2kPqvN--LGsmNaag-I59s8A2n_jFOKC7sha5cCLj8w2U4kO197BOD7Rlg5JNbAt7gpXSkm_0EXnfH7VWg4i25COV8sV8FfEgQlsAfXu5U_tJ7kSsSFH9CH-KwAa0zGPtXJ3P5TU20G3iNmSciXZVs9I3cv-K_aN8nw7IqjpXrh7yrcH86cnGhCLLDDtcXCNa6OKvttWNmiw4NSsX0OHR7zxt8IBQrmXzdXWwPAXRKvWdClTTbOmb257RyMFShx8su43smFBPx0TR2X5cdI_9I8LNI8rr0RegRkylYllj33GKUtH163FiiSUGef-S2C3baGVEd9XIDFva61Gdh8EGtBq8wEF0dSz4KCkmmWrNxeMEMBeJCoOIVq1uQYuQ3ITde8S0oi3zRs1ADFQ9jx5EgQdbhjLBumVzYW2pn2dzbYlFkFqMFNDM3xxMRxYb4IkslnC9Jjn2Iizn9CW3G02OnAoEuDE-GWbG9-DUx047K8BVhaSHt3p0O_rTS6nMYuLb90s15D7e2MTkK0ElbSUVyFaR3fV7dL56Q5dUhN26qVKdhHHLfqvRGZgW-aXwkiZiaNpIBBWfDfn-em_V5MTVryXq5I2qREyp_s_p-OTxUtdGoWH37rqsKTtWDhu-X36QXNASyYphFoVd98s0241kfNcnWRORnX8J8DU7SzL5tOeBoMMZoOiyBKhOBf4D41IAEqCWV1LXdlc3QtMjD1CQ.m3u8
#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID="720p60",NAME="720p60",AUTOSELECT=YES,DEFAULT=YES
#EXT-X-STREAM-INF:BANDWIDTH=3422999,RESOLUTION=1280x720,CODECS="avc1.4D401F,mp4a.40.2",VIDEO="720p60",FRAME-RATE=60.000
https://video-weaver.vie02.hls.ttvnw.net/v1/playlist/CucER7Nx8-bDWhB1j6v-CteThoTuBMmnFwWi12iyGCi8ki0E99WeO1Bkojmb3or2IrLQuT9IUXAWIi6-5wB32Ld4q6bVG5_mWZuzJZJLfgkJWLTeoBvz_jmlWbBQri5V3ZXeyCzm8Gh4X0aICo8bRqDuFld3F8PP1oQQK7WDd9Ar_bNkdVvBFgPSUcdLicl8AeWJuur_2TYVqBzVlnMsoGOIueim7Lll9ULYdD2OIF_c2di7zd44ZJHNd10rjcOmqwBjUOw4dyrIBt2P8Id1toeQ7KdN6K-M4e5mRH9gWf-pytQl3yguUiPp_XDdPQT_MbJKN92BKsbSlOvC4YoIlj4GhHuhXD7PsS1IOvOmXLUpn4cG8YXEXC-EMx0oqdc_8SazcM4eFzh3erDkX0g3M7W6r6dGo48nJ-LP5sjhjTwL1bjck2G08nX8YWo7lb6_hdDvpbGjUxmmSBhNN8-sWgVV1wlG7PcgZCKfDkKA8jAEi5hIUqV8HPoSdaT12QO8jck14N4nKjl9eOdTA5jw9yiZ3osZjHdFNDB5jQHEYc1KWnxJYI9P7iCmfIdaF7AK-1jhJmt7ys_RE5nnXQe6MyAUKs8iDfHFDIMLH77bKeqHMPQ4hyqiDpEAhEwrdA-WJ_-w1s-R3N1Bnj7wbvl5AlCp1f7Rh634ojVQvp_i19Uubae0bqzCP8WcboUo0h_wWf_pIu5X5xOIDjsSDxPzl42bCglPGoXNsiN8VWWBUcmbHHK3-ywteoLZ9ZzFwVEVfb1KMW5oV6ZF5BTScH1EDEQrWeDxy359Y6Z8uHSABaxAUvjwo_RNxueXGgwRL8QNbS1_xjMw044gASoJZXUtd2VzdC0yMPUJ.m3u8
#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID="720p30",NAME="720p",AUTOSELECT=YES,DEFAULT=YES
#EXT-X-STREAM-INF:BANDWIDTH=2373000,RESOLUTION=1280x720,CODECS="avc1.4D401F,mp4a.40.2",VIDEO="720p30",FRAME-RATE=30.000
https://video-weaver.vie02.hls.ttvnw.net/v1/playlist/Ct8Eg9po7IhRz_-WJyr8ky_iXF7q3Ck4j-9uQOJ9o65ukNqiJtM9D-VEnJ9KLCi40TasU0D6kzLAZpT5UYZbeNWPUOXjM-UkYogPJ_YNx0x_uj7JFXgUa-yLmcodGp0oL8059TVxIqe2YrJGg5EPDZX0cV7gB7eQ8-Vm42ClbHj40xlke99XYa29akuxFHadLgiULY9gk5YGiIzJ3gPezmB6BTjzKAqBIFh_hpMeKbl_1jyoVdNqEIOCaRhdayF1gxJCUEP8baYSCd3cClssChUW0YvC_ASlLcnxtODQ5Hjt8WpqT8AlwxVxkBAUe0uM5JJUCqYZqur-8pkZMpuR_lS8ZQ2xEOas-oj0PCJ55MGszF99_0aEdBb0NVPwuWxEuT0H5Ueel_Vr2xtokyokB45PmyUfug-GeCdKwEp9l46fApLna6wSUnh3bKFYy7YOduFWw8zc3Z2ly6i5LIkdaWI0-y4y4AdYEC7g41Fbt0xFmsG3mPa4r0BTCqJ6UsLuofur1moV30XwbtoOUOUbc3gZyYXoN9ZFHjHUNCW8gIvWkPw7iBG1VXUsmDqGQ1aUDgojMX-K8Qm5SjGKP-CYuPZ4urNygutacY-8p7GqxmjB20JE2zyuYz0CxFALAICS3MQc-Ib-MEFBf299a_gMv8pNw_X4xlQ-s0veHUu0inKUaxj8XUMWY7BRmmvdwE2R8NjgAMU0JDg1D4LTsILNurP9-6GqY9e9xEueynE-REFm415emR9JabulHwN-t78wgFuP1DLnqo8j4bvld-U35weiVkaqdGuEvgnK3cTWLKf_GxoM6HP-PHQyq0XqrSBgIAEqCWV1LXdlc3QtMjD1CQ.m3u8
#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID="480p30",NAME="480p",AUTOSELECT=YES,DEFAULT=YES
#EXT-X-STREAM-INF:BANDWIDTH=1427999,RESOLUTION=852x480,CODECS="avc1.4D401F,mp4a.40.2",VIDEO="480p30",FRAME-RATE=30.000
https://video-weaver.vie02.hls.ttvnw.net/v1/playlist/CtcECUByxqykyXEye1iOYfiH_L6MBBZOaqEu3r0jIUJOoA8gth_shx775gByxZQEvIZ4nACjtg9C_n49OldMiZboZBzrkGEnLCPlnjLHpPJuU3G-766Qc12uooer4jT2ZBObPWTp4yDYhBW7coIjc42rQ_nbHZegROOxYRi4oAf8ggOHgrN58I0Dg2T7Bv5yfIXCxfWxcDdsvGkrklcVhjQGWiTtm2UlZ_0zdzrKjHbc1BKFAG9i1LHzD6Uf4UH_ZzAWy7GdRT3rw0OkHXcuUyRD76fN4cbQUk830QKYGNeblEX0ustSlsj09_G2PS_kawjNGNCCSG79clEFK_a1-j5CXbWnXxo906RjQwJp4ij1ujwldvccQkPk8JbrHxEiVIQnvthFo5sFcmz07hdfwbNNUl4QgT7O5t7IUzUWQsqPPaSbaIy8hF8UskZ_V5nN_w116EQRLW2tz1vNGfNca2IYEmhWQd2GRFXdmA4ao7e8zZ2NU4YOtbh0XNubb7TkfZTw3ZBZ9d3wKCdJDjn7uplULmPFlLAqWATBGWNQoJiP0f3kAoDNcb2hM5NdYtF7kZeaMFw2woHFe3zaGPLmecC_uAWV7j7BJICKLVUvkNvjnknxlsTu-Hdwoi4uUML6cbyymTAAjQm1ZXu_kvYHanXNHVniQ585vBb5Omsj0Pf-j1i-tGXRbZ-8QrZO1ZfnzscKR_Rcmyoz3yg_6aosCOdNf6SPN3lXmaHt1hGovCCm6sH5MPJaTLukHtFKwDryTDDHxg7VucAoRSvtZFd0-IC132sGIPwjVq8aDBTXKvIT-D3T8PugkSABKglldS13ZXN0LTIw9Qk.m3u8
#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID="360p30",NAME="360p",AUTOSELECT=YES,DEFAULT=YES
#EXT-X-STREAM-INF:BANDWIDTH=630000,RESOLUTION=640x360,CODECS="avc1.4D401F,mp4a.40.2",VIDEO="360p30",FRAME-RATE=30.000
https://video-weaver.vie02.hls.ttvnw.net/v1/playlist/Cs8EXTwIYSW3SmlFFPLqx1286jP0YK21LS-B8ekCKyAhyBCS-Uqe6MnYSKMsohfOEqDXF9ix-k-O6OwFFsIj8WX9Ewq5FBisydjpqyMsbZYg3QTs_EiNy1-F6IBrqmfVfCfWvMVegPrk8TsFXIHu6WnXtlujrxfEtkGoKHZF7-LOjtILHpxBNqySey88AJk-bumgQnYor9on7BbtG1qpPurkovRM3r2A8n3_kbiu_Ij4sEbc2ZkpkZ2sK88kWO3uON-Qm3xyCF1zU1C2HZceo-r7a9_XLJCR_QXaDwpjDJHlEbYD-UAVnP1bNRCWR9pylSAuDAIFPkoB5bh06_an9tfB8NCupjaQRK56sxSLPpVDFUBFPhukpd2uVcGtFaJK_ezY2Q4G4HTZMBEbGMvZ4wjsHNoO3ZbhhcTN31HIm32G9uJAKaA6tQKzfsADJ-o0GGgNzdcRD4Wa181qzlPWggJOIBCn8EZr46Y2TLNpGB7vvXmf_aU4A6o9oo_ERB6v1_NgO1SmM_nDW31Tu7X4y4IuqK0IwqUFVSkhZyl9pz_OVxtIw5RS6CjP8NEt8E5KgshUwkx7ubYEn7HEDfsbH1atBKcd9uJFY9hIUw_gkFUT2QvVIqVcys8Pz7DfHtsWwnutdXTEa1AqxypUzHCxm7Lv4HWMWnoO1cqZ6BmaqeNsxoeH7qqodpGMSneDTNZaXttds17BakDflGOfXvvZYLNJnuVk-ki-CGXdgvtr1ic9XCqeqotbv-E23MgnRdI4qWoCOZi0mhpZ2qP68MGUCGJFGgx6-6uSVMdbrWbw9vUgASoJZXUtd2VzdC0yMPUJ.m3u8
#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID="160p30",NAME="160p",AUTOSELECT=YES,DEFAULT=YES
#EXT-X-STREAM-INF:BANDWIDTH=230000,RESOLUTION=284x160,CODECS="avc1.4D401F,mp4a.40.2",VIDEO="160p30",FRAME-RATE=30.000
https://video-weaver.vie02.hls.ttvnw.net/v1/playlist/CscE-2Ccq1atkR13_RxQvmky6mKGEplOsa7r_W_-Yz03b43gFACFhDo2UjfzPseyAyivkPD78wQb4s8wioLpqSDFodBTjKQ8ONRevu7XCmF7Iy_nZfB9QigIR6qUOyym3jdKnq-5QB_R9DHwJUAa3hoivIkcpJmZ0YAPVfuvdCY0flw2PrUBxgIBc7hJmCr-XdSq0Nl7SDhyxBCrZhmzsQk32TeCK96V4WJ880nbcOsfC1GI9SY0trMOUV6SY16z0Dkl38RfxCtGMjrtE9I7ZzL1gZU675_Wz_3LkOcWYfzvE3-OBTHk71ig2zUcXPTAyGxaStxAUhjj-TS21mfBcksa_6VVdKf8YGn8n7iO9m98lPuNZJ34AUU-0f3fjgn-2wOejl2FxJBoqXvEptu8el5nrNI0MfhVUdoNX9Vw30Q_XzK3pKrrZVtmWs6W38Y6jtuKm2jX_FHvSlfYLyPU94ZI9zYtCYAU0q8WDHwqJiPoHbA2DcwrnXw0viscEgDW4lV8TxfSJPA4F1YnahDlNLd0S4EcdtY2thFXR4PMerp4VgQymT1E38eMchlAVXSNG6XVQFP3J9F2Kw-j6LSlGyPy_4IO0OocPFYdC7EtTuy2jihzzZ6tbwkoESYLzgSI_M7iKMNyOYvZ46YYBHHEv-YMVa5vtSfSUtLWChiJFtJjHvrJHN5iwjcnZ47SJ9fnZ__5xeGo-Dsluu-GACpA2NpHCXGr7L6awQdsawSTUzZClFBLyXiapuBihPwBhyiBIav8H5wJs4OX8hoMS7aelbScB9r_PTOAIAEqCWV1LXdlc3QtMjD1CQ.m3u8
#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID="audio_only",NAME="audio_only",AUTOSELECT=NO,DEFAULT=NO
#EXT-X-STREAM-INF:BANDWIDTH=160000,CODECS="mp4a.40.2",VIDEO="audio_only"
https://video-weaver.vie02.hls.ttvnw.net/v1/playlist/Cr0EMO0S4cfoU9MnT7lbfkVp71n-M3rOvHB0kLQfaAl6uT3Z_iQRw_mQbOSl-KvXeOHeQrF7raeB4hJiSd9hqP86XW0lZpfT1466HAhRh8RhaeVzpX25AJSPtydBeDSYJpv08JyFkguZiWzuSb5DMcwMreqjflfltQeV0OouEbHzCIUFWEjA_mrdr4VNitHvQ5kI8iWzt0uL1ARwoHJUwlwByfBS-XS9MWtCvgjis6e-TnjT-wc38bOHcCl_gwrEZQcux4gfFrf1Kx37gAWSWXSJGF5MQZFrafZ4xrdQ73yFrL69mhJ-oyyopUxDd8GBc2iiwZTWqsBxRzo7F5yGCrcP03lI45OkVEMJlRiPG8TeQdx8c7y9FtiScmyAz6MY5yY7cVjyD4LRCyTPSro_DC9jcqncZ-OqhaWnolLy0gE29h4gYuMSoegm6vnDzNiHD2uxQeDJESYXK8_k19RcTSY3yaiASGIKEoRpLLiHTk6NMoubD3iZkmaYb-RYeRqZY_sg7focyr2fHJcD6kabhf7xXQ_viBsIgO9Q4CsvsXdcRjp0AKxxDiAMWC_miFC6c6_BW3auhLv02aNUlptXiZUqNJeEXin2ubVkdcQ9dHx26vc43ciYVoPhAY2sBzcMr1gpPRnkQK6izODqjJ1nXJAT28Qed5Cof3FV2nECrRuB5urgmCr_4zyOXO9dmR2qZ3GpPGnIQQtjoO87U2-lKTG0403fJ3okN23SwABMdqDRePCbISM_2eK5Xcq61ZWaGgyL6UGc6uBhFwOaB4sgASoJZXUtd2VzdC0yMPUJ.m3u8`
)

type List struct {
	Bandwidth  string `json:"BANDWIDTH"`
	Resolution string `json:"RESOLUTION"`
	Video      string `json:"VIDEO"`
	FrameRate  string `json:"FRAME-RATE"`
	URL        string
}

type MasterPlaylist struct {
	UsherURL string
	Lists    []List
}

func extractSegmentInfo(mediaParts []string, seg *List) {
	structType := reflect.TypeOf(*seg)
	structValue := reflect.ValueOf(seg).Elem()
	for _, part := range mediaParts {
		kv := strings.Split(part, "=")
		if len(kv) != 2 {
			continue
		}
		key := kv[0]
		value := kv[1]
		value, err := strconv.Unquote(value)
		if err != nil {
			value = kv[1]
		}

		for structId := 0; structId < structType.NumField()-1; structId++ {
			field := structType.Field(structId)
			tag := field.Tag.Get("json")
			if key == tag {
				structFied := structValue.FieldByName(field.Name)
				if structFied.IsValid() && structFied.CanSet() {
					structFied.SetString(value)
				}
				break
			}
		}
	}
}

func Parse(playlist string) MasterPlaylist {
	var master MasterPlaylist
	lines := strings.Split(playlist, "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "#EXT-X-STREAM-INF:") {
			var segment List
			segment.URL = lines[i+1]
			mediaParts := strings.Split(strings.Split(line, ":")[1], ",")
			extractSegmentInfo(mediaParts, &segment)
			master.Lists = append(master.Lists, segment)
		}
	}
	return master
}

func (playlist *MasterPlaylist) GetMediaPlaylist(quality string) (List, error) {
	segments := playlist.Lists
	for i := 0; i < len(segments); i++ {
		seg := segments[i]
		if quality == "best" && seg.Video == "chunked" {
			return seg, nil
		}
		if seg.Video == quality {
			return seg, nil
		}
	}
	return List{}, fmt.Errorf("could not find the provided quality for a livestream")
}

func (playlist *MasterPlaylist) GetJSONSegments() []string {
	var segments []string
	for _, seg := range playlist.Lists {
		b, err := json.MarshalIndent(seg, "", " ")
		if err != nil {
			break
		}
		segments = append(segments, string(b))
	}
	return segments
}
