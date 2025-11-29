// Package iata contains IATA airport codes, which are supported by the Google Flights API, along with time zones and coordinates.
// This package was generated using an airport list (which can be found at this address: [airports.json])
// and the Google Flights API.
//
// Command: go run ./iata/generate/generate.go
//
// Generation date: 2024-02-18
//
// [airports.json]: https://github.com/mwgg/Airports/blob/f259c38566a5acbcb04b64eb5ad01d14bf7fd07c/airports.json
package iata

// Location contains airport location data including city, timezone, and coordinates.
type Location struct {
	City string
	Tz   string
	Lat  float64
	Lon  float64
}

// IATATimeZone turns IATA airport codes into the time zone where the airport is located.
// If IATATimeZone can't find an IATA airport code, then it returns "Not supported IATA Code".
func IATATimeZone(iata string) Location {
	switch iata {
	case "AAA":
		return Location{"", "Pacific/Tahiti", -17.352600, -145.509995}
	case "AAE":
		return Location{"Annabah", "Africa/Algiers", 36.822201, 7.809170}
	case "AAL":
		return Location{"Aalborg", "Europe/Copenhagen", 57.092759, 9.849243}
	case "AAN":
		return Location{"Al Ain", "Asia/Dubai", 24.261700, 55.609200}
	case "AAP":
		return Location{"Samarinda", "Asia/Makassar", -0.373611, 117.255556}
	case "AAR":
		return Location{"Aarhus", "Europe/Copenhagen", 56.299999, 10.619000}
	case "AAT":
		return Location{"Altay", "Asia/Shanghai", 47.749886, 88.085808}
	case "AAX":
		return Location{"Araxa", "America/Sao_Paulo", -19.563200, -46.960400}
	case "AAZ":
		return Location{"Quezaltenango", "America/Guatemala", 14.865600, -91.501999}
	case "ABA":
		return Location{"Abakan", "Asia/Krasnoyarsk", 53.740002, 91.385002}
	case "ABB":
		return Location{"Asaba", "Africa/Lagos", 6.203333, 6.658889}
	case "ABD":
		return Location{"Abadan", "Asia/Tehran", 30.371099, 48.228298}
	case "ABE":
		return Location{"Allentown", "America/New_York", 40.652100, -75.440804}
	case "ABI":
		return Location{"Abilene", "America/Chicago", 32.411301, -99.681900}
	case "ABJ":
		return Location{"Abidjan", "Africa/Abidjan", 5.261390, -3.926290}
	case "ABL":
		return Location{"Ambler", "America/Anchorage", 67.106300, -157.856995}
	case "ABM":
		return Location{"", "Australia/Brisbane", -10.950800, 142.459000}
	case "ABQ":
		return Location{"Albuquerque", "America/Denver", 35.040199, -106.609001}
	case "ABR":
		return Location{"Aberdeen", "America/Chicago", 45.449100, -98.421799}
	case "ABS":
		return Location{"Abu Simbel", "Africa/Cairo", 22.375999, 31.611700}
	case "ABT":
		return Location{"", "Asia/Riyadh", 20.296101, 41.634300}
	case "ABU":
		return Location{"Atambua-Timor Island", "Asia/Makassar", -9.073050, 124.904999}
	case "ABV":
		return Location{"Abuja", "Africa/Lagos", 9.006790, 7.263170}
	case "ABX":
		return Location{"Albury", "Australia/Melbourne", -36.067799, 146.957993}
	case "ABY":
		return Location{"Albany", "America/New_York", 31.535500, -84.194504}
	case "ABZ":
		return Location{"Aberdeen", "Europe/London", 57.201900, -2.197780}
	case "ACA":
		return Location{"Acapulco", "America/Mexico_City", 16.757099, -99.753998}
	case "ACC":
		return Location{"Accra", "Africa/Accra", 5.605190, -0.166786}
	case "ACE":
		return Location{"Lanzarote Island", "Atlantic/Canary", 28.945499, -13.605200}
	case "ACF":
		return Location{"Brisbane", "Australia/Brisbane", -27.570299, 153.007996}
	case "ACH":
		return Location{"Altenrhein", "Europe/Vienna", 47.485001, 9.560770}
	case "ACI":
		return Location{"Saint Anne", "Europe/Guernsey", 49.706100, -2.214720}
	case "ACK":
		return Location{"Nantucket", "America/New_York", 41.253101, -70.060204}
	case "ACT":
		return Location{"Waco", "America/Chicago", 31.611300, -97.230499}
	case "ACV":
		return Location{"Arcata/Eureka", "America/Los_Angeles", 40.978100, -124.109001}
	case "ACX":
		return Location{"Xingyi", "Asia/Shanghai", 25.086389, 104.959444}
	case "ACY":
		return Location{"Atlantic City", "America/New_York", 39.457600, -74.577202}
	case "ACZ":
		return Location{"", "Asia/Tehran", 31.098301, 61.543900}
	case "ADA":
		return Location{"Adana", "Europe/Istanbul", 36.982201, 35.280399}
	case "ADB":
		return Location{"Izmir", "Europe/Istanbul", 38.292400, 27.157000}
	case "ADD":
		return Location{"Addis Ababa", "Africa/Addis_Ababa", 8.977890, 38.799301}
	case "ADE":
		return Location{"Aden", "Asia/Aden", 12.829500, 45.028801}
	case "ADF":
		return Location{"Adiyaman", "Europe/Istanbul", 37.731400, 38.468899}
	case "ADK":
		return Location{"Adak Island", "America/Adak", 51.877998, -176.645996}
	case "ADL":
		return Location{"Adelaide", "Australia/Adelaide", -34.945000, 138.531006}
	case "ADQ":
		return Location{"Kodiak", "America/Anchorage", 57.750000, -152.494003}
	case "ADU":
		return Location{"Ardabil", "Asia/Tehran", 38.325699, 48.424400}
	case "ADZ":
		return Location{"San Andres", "America/Bogota", 12.583600, -81.711200}
	case "AEB":
		return Location{"Baise", "Asia/Shanghai", 23.720600, 106.959999}
	case "AEP":
		return Location{"Buenos Aires", "America/Argentina/Buenos_Aires", -34.559200, -58.415600}
	case "AER":
		return Location{"Sochi", "Europe/Moscow", 43.449902, 39.956600}
	case "AES":
		return Location{"Alesund", "Europe/Oslo", 62.562500, 6.119700}
	case "AET":
		return Location{"Allakaket", "America/Anchorage", 66.551804, -152.621994}
	case "AEX":
		return Location{"Alexandria", "America/Chicago", 31.327400, -92.549797}
	case "AEY":
		return Location{"Akureyri", "Atlantic/Reykjavik", 65.660004, -18.072701}
	case "AFA":
		return Location{"San Rafael", "America/Argentina/Mendoza", -34.588299, -68.403900}
	case "AFL":
		return Location{"Alta Floresta", "America/Cuiaba", -9.866389, -56.105000}
	case "AGA":
		return Location{"Agadir", "Africa/Casablanca", 30.325001, -9.413070}
	case "AGH":
		return Location{"Angelholm", "Europe/Stockholm", 56.296101, 12.847100}
	case "AGP":
		return Location{"Malaga", "Europe/Madrid", 36.674900, -4.499110}
	case "AGR":
		return Location{"", "Asia/Kolkata", 27.155800, 77.960899}
	case "AGS":
		return Location{"Augusta", "America/New_York", 33.369900, -81.964500}
	case "AGT":
		return Location{"Ciudad del Este", "America/Asuncion", -25.459999, -54.840000}
	case "AGU":
		return Location{"Aguascalientes", "America/Mexico_City", 21.705601, -102.318001}
	case "AGV":
		return Location{"Acarigua", "America/Caracas", 9.553375, -69.237869}
	case "AGX":
		return Location{"", "Asia/Kolkata", 10.823700, 72.176003}
	case "AHB":
		return Location{"Abha", "Asia/Riyadh", 18.240400, 42.656601}
	case "AHE":
		return Location{"Ahe Atoll", "Pacific/Tahiti", -14.428100, -146.257004}
	case "AHO":
		return Location{"Alghero", "Europe/Rome", 40.632099, 8.290770}
	case "AHU":
		return Location{"Al Hoceima", "Africa/Casablanca", 35.177101, -3.839520}
	case "AIA":
		return Location{"Alliance", "America/Denver", 42.053200, -102.804001}
	case "AIN":
		return Location{"Wainwright", "America/Anchorage", 70.638000, -159.994995}
	case "AIR":
		return Location{"Paracatu", "America/Sao_Paulo", -17.036943, -46.260277}
	case "AIT":
		return Location{"Aitutaki", "Pacific/Rarotonga", -18.830900, -159.764008}
	case "AJA":
		return Location{"Ajaccio/Napoleon Bonaparte", "Europe/Paris", 41.923599, 8.802920}
	case "AJF":
		return Location{"Al-Jawf", "Asia/Riyadh", 29.785101, 40.099998}
	case "AJI":
		return Location{"Agri", "Europe/Istanbul", 39.654541, 43.025978}
	case "AJL":
		return Location{"Aizawl", "Asia/Kolkata", 23.840599, 92.619698}
	case "AJN":
		return Location{"Ouani", "Indian/Comoro", -12.131700, 44.430302}
	case "AJR":
		return Location{"Arvidsjaur", "Europe/Stockholm", 65.590302, 19.281900}
	case "AJU":
		return Location{"Aracaju", "America/Maceio", -10.984000, -37.070301}
	case "AKA":
		return Location{"Ankang", "Asia/Shanghai", 32.708099, 108.931000}
	case "AKB":
		return Location{"Atka", "America/Adak", 52.220299, -174.205994}
	case "AKI":
		return Location{"Akiak", "America/Anchorage", 60.902901, -161.231003}
	case "AKJ":
		return Location{"Asahikawa", "Asia/Tokyo", 43.670799, 142.447006}
	case "AKK":
		return Location{"Akhiok", "America/Anchorage", 56.938702, -154.182999}
	case "AKL":
		return Location{"Auckland", "Pacific/Auckland", -37.008099, 174.792007}
	case "AKN":
		return Location{"King Salmon", "America/Anchorage", 58.676800, -156.649002}
	case "AKP":
		return Location{"Anaktuvuk Pass", "America/Anchorage", 68.133598, -151.742996}
	case "AKR":
		return Location{"Akure", "Africa/Lagos", 7.246740, 5.301010}
	case "AKS":
		return Location{"Auki", "Pacific/Guadalcanal", -8.702570, 160.682007}
	case "AKU":
		return Location{"Aksu", "Asia/Shanghai", 41.262501, 80.291702}
	case "AKV":
		return Location{"Akulivik", "America/Iqaluit", 60.818600, -78.148598}
	case "AKX":
		return Location{"Aktyubinsk", "Asia/Aqtobe", 50.245800, 57.206699}
	case "AKY":
		return Location{"Sittwe", "Asia/Yangon", 20.132700, 92.872597}
	case "ALA":
		return Location{"Almaty", "Asia/Almaty", 43.352100, 77.040497}
	case "ALB":
		return Location{"Albany", "America/New_York", 42.748299, -73.801697}
	case "ALC":
		return Location{"Alicante", "Europe/Madrid", 38.282200, -0.558156}
	case "ALF":
		return Location{"Alta", "Europe/Oslo", 69.976097, 23.371700}
	case "ALG":
		return Location{"Algiers", "Africa/Algiers", 36.691002, 3.215410}
	case "ALH":
		return Location{"Albany", "Australia/Perth", -34.943298, 117.808998}
	case "ALI":
		return Location{"Alice", "America/Chicago", 27.740900, -98.026901}
	case "ALO":
		return Location{"Waterloo", "America/Chicago", 42.557098, -92.400299}
	case "ALP":
		return Location{"Aleppo", "Asia/Damascus", 36.180698, 37.224400}
	case "ALQ":
		return Location{"Alegrete", "America/Sao_Paulo", -29.812700, -55.893398}
	case "ALS":
		return Location{"Alamosa", "America/Denver", 37.434898, -105.866997}
	case "ALW":
		return Location{"Walla Walla", "America/Los_Angeles", 46.094898, -118.288002}
	case "AMA":
		return Location{"Amarillo", "America/Chicago", 35.219398, -101.706001}
	case "AMD":
		return Location{"Ahmedabad", "Asia/Kolkata", 23.077200, 72.634697}
	case "AMH":
		return Location{"", "Africa/Addis_Ababa", 6.039390, 37.590500}
	case "AMM":
		return Location{"Amman", "Asia/Amman", 31.722601, 35.993198}
	case "AMQ":
		return Location{"Ambon", "Asia/Jayapura", -3.710260, 128.089005}
	case "AMS":
		return Location{"Amsterdam", "Europe/Amsterdam", 52.308601, 4.763890}
	case "ANC":
		return Location{"Anchorage", "America/Anchorage", 61.174400, -149.996002}
	case "ANF":
		return Location{"Antofagasta", "America/Santiago", -23.444500, -70.445099}
	case "ANI":
		return Location{"Aniak", "America/Anchorage", 61.581600, -159.542999}
	case "ANR":
		return Location{"Antwerp", "Europe/Brussels", 51.189400, 4.460280}
	case "ANU":
		return Location{"St. George", "America/Antigua", 17.136700, -61.792702}
	case "ANV":
		return Location{"Anvik", "America/Anchorage", 62.646702, -160.190994}
	case "ANX":
		return Location{"Andenes", "Europe/Oslo", 69.292503, 16.144199}
	case "AOE":
		return Location{"Eskisehir", "Europe/Istanbul", 39.809898, 30.519400}
	case "AOG":
		return Location{"Anshan", "Asia/Shanghai", 41.105301, 122.853996}
	case "AOI":
		return Location{"Ancona", "Europe/Rome", 43.616299, 13.362300}
	case "AOJ":
		return Location{"Aomori", "Asia/Tokyo", 40.734699, 140.690994}
	case "AOK":
		return Location{"Karpathos Island", "Europe/Athens", 35.421398, 27.146000}
	case "AOO":
		return Location{"Altoona", "America/New_York", 40.296398, -78.320000}
	case "AOR":
		return Location{"Alor Satar", "Asia/Kuala_Lumpur", 6.189670, 100.398003}
	case "APK":
		return Location{"Apataki", "Pacific/Tahiti", -15.573600, -146.414993}
	case "APL":
		return Location{"Nampula", "Africa/Maputo", -15.105600, 39.281799}
	case "APN":
		return Location{"Alpena", "America/Detroit", 45.078098, -83.560303}
	case "APO":
		return Location{"Carepa", "America/Bogota", 7.811960, -76.716400}
	case "APW":
		return Location{"Apia", "Pacific/Apia", -13.830000, -172.007996}
	case "AQG":
		return Location{"Anqing", "Asia/Shanghai", 30.582199, 117.050003}
	case "AQI":
		return Location{"Qaisumah", "Asia/Riyadh", 28.335199, 46.125099}
	case "AQJ":
		return Location{"Aqaba", "Asia/Amman", 29.611601, 35.018101}
	case "AQP":
		return Location{"Arequipa", "America/Lima", -16.341101, -71.583099}
	case "ARC":
		return Location{"Arctic Village", "America/Anchorage", 68.114700, -145.578995}
	case "ARD":
		return Location{"Alor Island", "Asia/Makassar", -8.132340, 124.597000}
	case "ARH":
		return Location{"Archangelsk", "Europe/Moscow", 64.600304, 40.716702}
	case "ARI":
		return Location{"Arica", "America/Lima", -18.348499, -70.338699}
	case "ARK":
		return Location{"Arusha", "Africa/Dar_es_Salaam", -3.367790, 36.633301}
	case "ARM":
		return Location{"Armidale", "Australia/Sydney", -30.528099, 151.617004}
	case "ARN":
		return Location{"Stockholm", "Europe/Stockholm", 59.651901, 17.918600}
	case "ART":
		return Location{"Watertown", "America/New_York", 43.991901, -76.021698}
	case "ARU":
		return Location{"Aracatuba", "America/Sao_Paulo", -21.141300, -50.424702}
	case "ASB":
		return Location{"Ashgabat", "Asia/Ashgabat", 37.986801, 58.361000}
	case "ASE":
		return Location{"Aspen", "America/Denver", 39.223202, -106.869003}
	case "ASF":
		return Location{"Astrakhan", "Europe/Astrakhan", 46.283298, 48.006302}
	case "ASJ":
		return Location{"Amami", "Asia/Tokyo", 28.430599, 129.712997}
	case "ASM":
		return Location{"Asmara", "Africa/Asmara", 15.291900, 38.910702}
	case "ASO":
		return Location{"Asosa", "Africa/Addis_Ababa", 10.018500, 34.586300}
	case "ASP":
		return Location{"Alice Springs", "Australia/Darwin", -23.806700, 133.901993}
	case "ASR":
		return Location{"Kayseri", "Europe/Istanbul", 38.770401, 35.495399}
	case "ASU":
		return Location{"Asuncion", "America/Asuncion", -25.240000, -57.520000}
	case "ASV":
		return Location{"Amboseli National Park", "Africa/Nairobi", -2.645050, 37.253101}
	case "ASW":
		return Location{"Aswan", "Africa/Cairo", 23.964399, 32.820000}
	case "ATA":
		return Location{"Anta", "America/Lima", -9.347440, -77.598396}
	case "ATD":
		return Location{"Atoifi", "Pacific/Guadalcanal", -8.873330, 161.011002}
	case "ATH":
		return Location{"Athens", "Europe/Athens", 37.936401, 23.944500}
	case "ATK":
		return Location{"Atqasuk", "America/Anchorage", 70.467300, -157.436005}
	case "ATL":
		return Location{"Atlanta", "America/New_York", 33.636700, -84.428101}
	case "ATM":
		return Location{"Altamira", "America/Santarem", -3.253910, -52.254002}
	case "ATQ":
		return Location{"Amritsar", "Asia/Kolkata", 31.709600, 74.797302}
	case "ATW":
		return Location{"Appleton", "America/Chicago", 44.258099, -88.519096}
	case "ATY":
		return Location{"Watertown", "America/Chicago", 44.914001, -97.154701}
	case "ATZ":
		return Location{"Assiut", "Africa/Cairo", 27.046499, 31.011999}
	case "AUA":
		return Location{"Oranjestad", "America/Aruba", 12.501400, -70.015198}
	case "AUC":
		return Location{"Arauca", "America/Bogota", 7.068880, -70.736900}
	case "AUG":
		return Location{"Augusta", "America/New_York", 44.320599, -69.797302}
	case "AUH":
		return Location{"Abu Dhabi", "Asia/Dubai", 24.433001, 54.651100}
	case "AUK":
		return Location{"Alakanuk", "America/Nome", 62.680042, -164.659927}
	case "AUQ":
		return Location{"", "Pacific/Marquesas", -9.768790, -139.011002}
	case "AUR":
		return Location{"Aurillac", "Europe/Paris", 44.891399, 2.421940}
	case "AUS":
		return Location{"Austin", "America/Chicago", 30.194500, -97.669899}
	case "AUU":
		return Location{"", "Australia/Brisbane", -13.353900, 141.720993}
	case "AUX":
		return Location{"Araguaina", "America/Araguaina", -7.227870, -48.240501}
	case "AUY":
		return Location{"Anelghowhat", "Pacific/Efate", -20.249201, 169.770996}
	case "AVA":
		return Location{"Anshun", "Asia/Shanghai", 26.260556, 105.873333}
	case "AVL":
		return Location{"Asheville", "America/New_York", 35.436199, -82.541801}
	case "AVP":
		return Location{"Wilkes-Barre/Scranton", "America/New_York", 41.338501, -75.723396}
	case "AVV":
		return Location{"Melbourne", "Australia/Melbourne", -38.039398, 144.468994}
	case "AWA":
		return Location{"Awassa", "Africa/Addis_Ababa", 7.067000, 38.500000}
	case "AWD":
		return Location{"Aniwa", "Pacific/Efate", -19.240000, 169.604996}
	case "AWZ":
		return Location{"Ahwaz", "Asia/Tehran", 31.337400, 48.762001}
	case "AXA":
		return Location{"The Valley", "America/Anguilla", 18.204800, -63.055099}
	case "AXD":
		return Location{"Alexandroupolis", "Europe/Athens", 40.855900, 25.956301}
	case "AXJ":
		return Location{"", "Asia/Tokyo", 32.482498, 130.158997}
	case "AXM":
		return Location{"Armenia", "America/Bogota", 4.452780, -75.766400}
	case "AXP":
		return Location{"Spring Point", "America/Nassau", 22.441799, -73.970901}
	case "AXR":
		return Location{"", "Pacific/Tahiti", -15.248300, -146.617004}
	case "AXT":
		return Location{"Akita", "Asia/Tokyo", 39.615601, 140.218994}
	case "AYP":
		return Location{"Ayacucho", "America/Lima", -13.154800, -74.204399}
	case "AYQ":
		return Location{"Ayers Rock", "Australia/Darwin", -25.186100, 130.975998}
	case "AYT":
		return Location{"Antalya", "Europe/Istanbul", 36.898701, 30.800501}
	case "AZA":
		return Location{"Phoenix", "America/Phoenix", 33.307800, -111.654999}
	case "AZD":
		return Location{"Yazd", "Asia/Tehran", 31.904900, 54.276501}
	case "AZI":
		return Location{"", "Asia/Dubai", 24.428301, 54.458099}
	case "AZO":
		return Location{"Kalamazoo", "America/Detroit", 42.234901, -85.552101}
	case "AZR":
		return Location{"", "Africa/Algiers", 27.837601, -0.186414}
	case "AZS":
		return Location{"Samana", "America/Santo_Domingo", 19.267000, -69.741997}
	case "BAG":
		return Location{"Baguio City", "Asia/Manila", 16.375099, 120.620003}
	case "BAH":
		return Location{"Manama", "Asia/Bahrain", 26.270800, 50.633598}
	case "BAL":
		return Location{"Batman", "Europe/Istanbul", 37.929001, 41.116600}
	case "BAQ":
		return Location{"Barranquilla", "America/Bogota", 10.889600, -74.780800}
	case "BAS":
		return Location{"Ballalae", "Pacific/Guadalcanal", -6.990745, 155.886656}
	case "BAU":
		return Location{"Bauru", "America/Sao_Paulo", -22.344999, -49.053799}
	case "BAV":
		return Location{"Baotou", "Asia/Shanghai", 40.560001, 109.997002}
	case "BAX":
		return Location{"Barnaul", "Asia/Barnaul", 53.363800, 83.538498}
	case "BAY":
		return Location{"Baia Mare", "Europe/Bucharest", 47.658401, 23.469999}
	case "BAZ":
		return Location{"Barcelos", "America/Manaus", -0.981292, -62.919601}
	case "BBA":
		return Location{"Balmaceda", "America/Santiago", -45.916100, -71.689499}
	case "BBI":
		return Location{"Bhubaneswar", "Asia/Kolkata", 20.244400, 85.817802}
	case "BBK":
		return Location{"Kasane", "Africa/Gaborone", -17.832899, 25.162399}
	case "BBN":
		return Location{"Bario", "Asia/Kuching", 3.733890, 115.478996}
	case "BBU":
		return Location{"Bucharest", "Europe/Bucharest", 44.503201, 26.102100}
	case "BCD":
		return Location{"Bacolod City", "Asia/Manila", 10.776400, 123.014999}
	case "BCI":
		return Location{"Barcaldine", "Australia/Brisbane", -23.565300, 145.307007}
	case "BCM":
		return Location{"Bacau", "Europe/Bucharest", 46.521900, 26.910299}
	case "BCN":
		return Location{"Barcelona", "Europe/Madrid", 41.297100, 2.078460}
	case "BCO":
		return Location{"Baco", "Africa/Addis_Ababa", 5.782870, 36.562000}
	case "BCT":
		return Location{"Boca Raton", "America/New_York", 26.378500, -80.107697}
	case "BDA":
		return Location{"Hamilton", "Atlantic/Bermuda", 32.363998, -64.678703}
	case "BDB":
		return Location{"Bundaberg", "Australia/Brisbane", -24.903900, 152.319000}
	case "BDD":
		return Location{"", "Australia/Brisbane", -10.150000, 142.173400}
	case "BDJ":
		return Location{"Banjarmasin-Borneo Island", "Asia/Makassar", -3.442360, 114.763000}
	case "BDL":
		return Location{"Hartford", "America/New_York", 41.938900, -72.683197}
	case "BDO":
		return Location{"Bandung-Java Island", "Asia/Jakarta", -6.900630, 107.575996}
	case "BDP":
		return Location{"Bhadrapur", "Asia/Kathmandu", 26.570801, 88.079597}
	case "BDQ":
		return Location{"Vadodara", "Asia/Kolkata", 22.336201, 73.226303}
	case "BDR":
		return Location{"Bridgeport", "America/New_York", 41.163502, -73.126198}
	case "BDS":
		return Location{"Brindisi", "Europe/Rome", 40.657600, 17.947001}
	case "BDU":
		return Location{"Malselv", "Europe/Oslo", 69.055801, 18.540400}
	case "BEB":
		return Location{"Balivanich", "Europe/London", 57.481098, -7.362780}
	case "BEG":
		return Location{"Belgrad", "Europe/Belgrade", 44.818401, 20.309099}
	case "BEJ":
		return Location{"Tanjung Redep-Borneo Island", "Asia/Makassar", 2.155500, 117.431999}
	case "BEK":
		return Location{"", "Asia/Kolkata", 28.422100, 79.450798}
	case "BEL":
		return Location{"Belem", "America/Belem", -1.379250, -48.476299}
	case "BEM":
		return Location{"Bossembele", "Africa/Bangui", 5.267000, 17.632999}
	case "BEN":
		return Location{"Benghazi", "Africa/Tripoli", 32.096802, 20.269501}
	case "BER":
		return Location{"Berlin", "Europe/Berlin", 52.366667, 13.503333}
	case "BES":
		return Location{"Brest/Guipavas", "Europe/Paris", 48.447899, -4.418540}
	case "BET":
		return Location{"Bethel", "America/Anchorage", 60.779800, -161.837997}
	case "BEU":
		return Location{"", "Australia/Brisbane", -24.346100, 139.460007}
	case "BEW":
		return Location{"Beira", "Africa/Maputo", -19.796400, 34.907600}
	case "BEY":
		return Location{"Beirut", "Asia/Beirut", 33.820900, 35.488400}
	case "BFD":
		return Location{"Bradford", "America/New_York", 41.803101, -78.640099}
	case "BFF":
		return Location{"Scottsbluff", "America/Denver", 41.874001, -103.596001}
	case "BFI":
		return Location{"Seattle", "America/Los_Angeles", 47.529999, -122.302002}
	case "BFJ":
		return Location{"", "Pacific/Fiji", -17.535056, 177.680917}
	case "BFL":
		return Location{"Bakersfield", "America/Los_Angeles", 35.433601, -119.056999}
	case "BFM":
		return Location{"Mobile", "America/Chicago", 30.626801, -88.068100}
	case "BFN":
		return Location{"Bloemfontain", "Africa/Johannesburg", -29.092699, 26.302401}
	case "BFS":
		return Location{"Belfast", "Europe/London", 54.657501, -6.215830}
	case "BFV":
		return Location{"", "Asia/Bangkok", 15.229500, 103.252998}
	case "BFX":
		return Location{"Bafoussam", "Africa/Douala", 5.536920, 10.354600}
	case "BGA":
		return Location{"Bucaramanga", "America/Bogota", 7.126500, -73.184800}
	case "BGF":
		return Location{"Bangui", "Africa/Bangui", 4.398480, 18.518801}
	case "BGI":
		return Location{"Bridgetown", "America/Barbados", 13.074600, -59.492500}
	case "BGM":
		return Location{"Binghamton", "America/New_York", 42.208698, -75.979797}
	case "BGN":
		return Location{"Belaya Gora", "Asia/Magadan", 68.556944, 146.234722}
	case "BGO":
		return Location{"Bergen", "Europe/Oslo", 60.293400, 5.218140}
	case "BGR":
		return Location{"Bangor", "America/New_York", 44.807400, -68.828102}
	case "BGS":
		return Location{"Big Spring", "America/Chicago", 32.212601, -101.522003}
	case "BGW":
		return Location{"Baghdad", "Asia/Baghdad", 33.262501, 44.234600}
	case "BGX":
		return Location{"Bage", "America/Sao_Paulo", -31.390499, -54.112202}
	case "BGY":
		return Location{"Bergamo", "Europe/Rome", 45.673901, 9.704170}
	case "BHB":
		return Location{"Bar Harbor", "America/New_York", 44.450001, -68.361504}
	case "BHD":
		return Location{"Belfast", "Europe/London", 54.618099, -5.872500}
	case "BHE":
		return Location{"Blenheim", "Pacific/Auckland", -41.518299, 173.869995}
	case "BHH":
		return Location{"", "Asia/Riyadh", 19.984400, 42.620899}
	case "BHI":
		return Location{"Bahia Blanca", "America/Argentina/Buenos_Aires", -38.725000, -62.169300}
	case "BHJ":
		return Location{"Bhuj", "Asia/Kolkata", 23.287800, 69.670197}
	case "BHK":
		return Location{"Bukhara", "Asia/Samarkand", 39.775002, 64.483299}
	case "BHM":
		return Location{"Birmingham", "America/Chicago", 33.562901, -86.753502}
	case "BHO":
		return Location{"Bhopal", "Asia/Kolkata", 23.287500, 77.337402}
	case "BHQ":
		return Location{"Broken Hill", "Australia/Broken_Hill", -32.001400, 141.472000}
	case "BHR":
		return Location{"Bharatpur", "Asia/Kathmandu", 27.678101, 84.429398}
	case "BHU":
		return Location{"Bhavnagar", "Asia/Kolkata", 21.752199, 72.185204}
	case "BHX":
		return Location{"Birmingham", "Europe/London", 52.453899, -1.748030}
	case "BHY":
		return Location{"Beihai", "Asia/Shanghai", 21.539400, 109.293999}
	case "BIA":
		return Location{"Bastia/Poretta", "Europe/Paris", 42.552700, 9.483730}
	case "BIH":
		return Location{"Bishop", "America/Los_Angeles", 37.373100, -118.363998}
	case "BIK":
		return Location{"Biak-Supiori Island", "Asia/Jayapura", -1.190020, 136.108002}
	case "BIL":
		return Location{"Billings", "America/Denver", 45.807701, -108.542999}
	case "BIM":
		return Location{"South Bimini", "America/Nassau", 25.699900, -79.264702}
	case "BIO":
		return Location{"Bilbao", "Europe/Madrid", 43.301102, -2.910610}
	case "BIQ":
		return Location{"Biarritz/Anglet/Bayonne", "Europe/Paris", 43.468399, -1.523320}
	case "BIR":
		return Location{"Biratnagar", "Asia/Kathmandu", 26.481501, 87.264000}
	case "BIS":
		return Location{"Bismarck", "America/Chicago", 46.772701, -100.746002}
	case "BJA":
		return Location{"Bejaia", "Africa/Algiers", 36.712002, 5.069920}
	case "BJC":
		return Location{"Denver", "America/Denver", 39.908798, -105.116997}
	case "BJF":
		return Location{"Batsfjord", "Europe/Oslo", 70.600502, 29.691401}
	case "BJI":
		return Location{"Bemidji", "America/Chicago", 47.509399, -94.933701}
	case "BJL":
		return Location{"Banjul", "Africa/Banjul", 13.338000, -16.652201}
	case "BJM":
		return Location{"Bujumbura", "Africa/Bujumbura", -3.324020, 29.318501}
	case "BJR":
		return Location{"Bahir Dar", "Africa/Addis_Ababa", 11.608100, 37.321602}
	case "BJV":
		return Location{"Bodrum", "Europe/Istanbul", 37.250599, 27.664301}
	case "BJW":
		return Location{"Bajawa", "Asia/Makassar", -8.808140, 120.996002}
	case "BJX":
		return Location{"Silao", "America/Mexico_City", 20.993500, -101.481003}
	case "BJZ":
		return Location{"Badajoz", "Europe/Madrid", 38.891300, -6.821330}
	case "BKA":
		return Location{"Moscow", "Europe/Moscow", 55.617199, 38.060001}
	case "BKB":
		return Location{"Bikaner", "Asia/Kolkata", 28.070601, 73.207199}
	case "BKC":
		return Location{"Buckland", "America/Anchorage", 65.981598, -161.149002}
	case "BKG":
		return Location{"Branson", "America/Chicago", 36.532082, -93.200544}
	case "BKI":
		return Location{"Kota Kinabalu", "Asia/Kuching", 5.937210, 116.051003}
	case "BKK":
		return Location{"Bangkok", "Asia/Bangkok", 13.681100, 100.747002}
	case "BKM":
		return Location{"Bakalalan", "Asia/Kuching", 3.974000, 115.617996}
	case "BKO":
		return Location{"Senou", "Africa/Bamako", 12.533500, -7.949940}
	case "BKQ":
		return Location{"Blackall", "Australia/Brisbane", -24.427799, 145.429001}
	case "BKS":
		return Location{"Bengkulu-Sumatra Island", "Asia/Jakarta", -3.863700, 102.338997}
	case "BKW":
		return Location{"Beckley", "America/New_York", 37.787300, -81.124199}
	case "BKZ":
		return Location{"Bukoba", "Africa/Dar_es_Salaam", -1.332000, 31.821200}
	case "BLA":
		return Location{"Barcelona", "America/Caracas", 10.107100, -64.689201}
	case "BLB":
		return Location{"Panama City", "America/Panama", 8.914790, -79.599602}
	case "BLD":
		return Location{"Boulder City", "America/Los_Angeles", 35.947498, -114.861000}
	case "BLI":
		return Location{"Bellingham", "America/Los_Angeles", 48.792801, -122.538002}
	case "BLJ":
		return Location{"Batna", "Africa/Algiers", 35.752102, 6.308590}
	case "BLL":
		return Location{"Billund", "Europe/Copenhagen", 55.740299, 9.151780}
	case "BLQ":
		return Location{"Bologna", "Europe/Rome", 44.535400, 11.288700}
	case "BLR":
		return Location{"Bangalore", "Asia/Kolkata", 13.197900, 77.706299}
	case "BLV":
		return Location{"Belleville", "America/Chicago", 38.545200, -89.835197}
	case "BLZ":
		return Location{"Blantyre", "Africa/Blantyre", -15.679100, 34.973999}
	case "BMA":
		return Location{"Stockholm", "Europe/Stockholm", 59.354401, 17.941700}
	case "BME":
		return Location{"Broome", "Australia/Perth", -17.944700, 122.232002}
	case "BMI":
		return Location{"Bloomington-Normal", "America/Chicago", 40.477100, -88.915901}
	case "BMO":
		return Location{"Banmaw", "Asia/Yangon", 24.268999, 97.246201}
	case "BMU":
		return Location{"Bima-Sumbawa Island", "Asia/Makassar", -8.539650, 118.686996}
	case "BMV":
		return Location{"Buon Ma Thuot", "Asia/Ho_Chi_Minh", 12.668300, 108.120003}
	case "BMW":
		return Location{"Bordj Badji Mokhtar", "Africa/Algiers", 21.375000, 0.923889}
	case "BNA":
		return Location{"Nashville", "America/Chicago", 36.124500, -86.678200}
	case "BND":
		return Location{"Bandar Abbas", "Asia/Tehran", 27.218300, 56.377800}
	case "BNE":
		return Location{"Brisbane", "Australia/Brisbane", -27.384199, 153.117004}
	case "BNI":
		return Location{"Benin", "Africa/Lagos", 6.316980, 5.599500}
	case "BNK":
		return Location{"Ballina", "Australia/Sydney", -28.833900, 153.561996}
	case "BNN":
		return Location{"Bronnoy", "Europe/Oslo", 65.461098, 12.217500}
	case "BNS":
		return Location{"Barinas", "America/Caracas", 8.619570, -70.220802}
	case "BNX":
		return Location{"Banja Luka", "Europe/Sarajevo", 44.941399, 17.297501}
	case "BNY":
		return Location{"Anua", "Pacific/Guadalcanal", -11.302222, 159.798333}
	case "BOB":
		return Location{"Motu Mute", "Pacific/Tahiti", -16.444401, -151.751007}
	case "BOC":
		return Location{"Isla Colon", "America/Panama", 9.340850, -82.250801}
	case "BOD":
		return Location{"Bordeaux/Merignac", "Europe/Paris", 44.828300, -0.715556}
	case "BOG":
		return Location{"Bogota", "America/Bogota", 4.701590, -74.146900}
	case "BOH":
		return Location{"Bournemouth", "Europe/London", 50.779999, -1.842500}
	case "BOI":
		return Location{"Boise", "America/Boise", 43.564400, -116.223000}
	case "BOJ":
		return Location{"Burgas", "Europe/Sofia", 42.569599, 27.515200}
	case "BOM":
		return Location{"Mumbai", "Asia/Kolkata", 19.088699, 72.867897}
	case "BON":
		return Location{"Kralendijk", "America/Kralendijk", 12.131000, -68.268501}
	case "BOO":
		return Location{"Bodo", "Europe/Oslo", 67.269203, 14.365300}
	case "BOS":
		return Location{"Boston", "America/New_York", 42.364300, -71.005203}
	case "BOY":
		return Location{"Bobo Dioulasso", "Africa/Ouagadougou", 11.160100, -4.330970}
	case "BPG":
		return Location{"Barra Do Garcas", "America/Cuiaba", -15.861300, -52.388901}
	case "BPL":
		return Location{"Bole", "Asia/Shanghai", 44.895000, 82.300000}
	case "BPN":
		return Location{"Balikpapan-Borneo Island", "Asia/Makassar", -1.268270, 116.893997}
	case "BPS":
		return Location{"Porto Seguro", "America/Bahia", -16.438601, -39.080898}
	case "BPT":
		return Location{"Beaumont/Port Arthur", "America/Chicago", 29.950800, -94.020699}
	case "BPX":
		return Location{"Bangda", "Asia/Shanghai", 30.553600, 97.108299}
	case "BQB":
		return Location{"Busselton", "Australia/Perth", -33.688423, 115.401596}
	case "BQD":
		return Location{"Budardalur", "Atlantic/Reykjavik", 65.075302, -21.800301}
	case "BQK":
		return Location{"Brunswick", "America/New_York", 31.258801, -81.466499}
	case "BQL":
		return Location{"", "Australia/Brisbane", -22.913300, 139.899994}
	case "BQN":
		return Location{"Aguadilla", "America/Puerto_Rico", 18.494900, -67.129402}
	case "BQS":
		return Location{"Blagoveschensk", "Asia/Yakutsk", 50.425400, 127.412003}
	case "BQW":
		return Location{"", "Australia/Perth", -20.148300, 127.973000}
	case "BRA":
		return Location{"Barreiras", "America/Bahia", -12.078900, -45.008999}
	case "BRB":
		return Location{"", "America/Fortaleza", -2.755556, -42.810000}
	case "BRC":
		return Location{"San Carlos de Bariloche", "America/Argentina/Salta", -41.151199, -71.157501}
	case "BRD":
		return Location{"Brainerd", "America/Chicago", 46.398300, -94.138100}
	case "BRE":
		return Location{"Bremen", "Europe/Berlin", 53.047501, 8.786670}
	case "BRI":
		return Location{"Bari", "Europe/Rome", 41.138901, 16.760599}
	case "BRK":
		return Location{"", "Australia/Sydney", -30.039200, 145.951996}
	case "BRL":
		return Location{"Burlington", "America/Chicago", 40.783199, -91.125504}
	case "BRM":
		return Location{"Barquisimeto", "America/Caracas", 10.042747, -69.358620}
	case "BRN":
		return Location{"Bern", "Europe/Zurich", 46.914101, 7.497150}
	case "BRO":
		return Location{"Brownsville", "America/Chicago", 25.906799, -97.425903}
	case "BRQ":
		return Location{"Brno", "Europe/Prague", 49.151299, 16.694401}
	case "BRR":
		return Location{"Eoligarry", "Europe/London", 57.022800, -7.443060}
	case "BRS":
		return Location{"Bristol", "Europe/London", 51.382702, -2.719090}
	case "BRU":
		return Location{"Brussels", "Europe/Brussels", 50.901402, 4.484440}
	case "BRW":
		return Location{"Barrow", "America/Anchorage", 71.285400, -156.766006}
	case "BSA":
		return Location{"Bosaso", "Africa/Mogadishu", 11.275300, 49.149399}
	case "BSB":
		return Location{"Brasilia", "America/Sao_Paulo", -15.869167, -47.920834}
	case "BSC":
		return Location{"Bahia Solano", "America/Bogota", 6.202920, -77.394700}
	case "BSD":
		return Location{"", "Asia/Shanghai", 25.053301, 99.168297}
	case "BSG":
		return Location{"", "Africa/Malabo", 1.905470, 9.805680}
	case "BSK":
		return Location{"Biskra", "Africa/Algiers", 34.793301, 5.738230}
	case "BSL":
		return Location{"Bale/Mulhouse", "Europe/Paris", 47.589600, 7.529910}
	case "BSO":
		return Location{"Basco", "Asia/Manila", 20.451300, 121.980003}
	case "BSR":
		return Location{"Basrah", "Asia/Baghdad", 30.549101, 47.662102}
	case "BTC":
		return Location{"Batticaloa", "Asia/Colombo", 7.705760, 81.678802}
	case "BTH":
		return Location{"Batam Island", "Asia/Jakarta", 1.121030, 104.119003}
	case "BTI":
		return Location{"Barter Island Lrrs", "America/Anchorage", 70.134003, -143.582001}
	case "BTJ":
		return Location{"Banda Aceh-Sumatra Island", "Asia/Jakarta", 5.523520, 95.420403}
	case "BTK":
		return Location{"Bratsk", "Asia/Irkutsk", 56.370602, 101.697998}
	case "BTM":
		return Location{"Butte", "America/Denver", 45.954800, -112.497002}
	case "BTR":
		return Location{"Baton Rouge", "America/Chicago", 30.533199, -91.149597}
	case "BTS":
		return Location{"Bratislava", "Europe/Bratislava", 48.170200, 17.212700}
	case "BTT":
		return Location{"Bettles", "America/Anchorage", 66.913902, -151.529007}
	case "BTU":
		return Location{"Bintulu", "Asia/Kuching", 3.123850, 113.019997}
	case "BTV":
		return Location{"Burlington", "America/New_York", 44.471901, -73.153297}
	case "BTW":
		return Location{"Batu Licin-Borneo Island", "Asia/Makassar", -3.412410, 115.995003}
	case "BUA":
		return Location{"Buka Island", "Pacific/Bougainville", -5.422320, 154.673004}
	case "BUC":
		return Location{"", "Australia/Brisbane", -17.748600, 139.533997}
	case "BUD":
		return Location{"Budapest", "Europe/Budapest", 47.436901, 19.255600}
	case "BUF":
		return Location{"Buffalo", "America/New_York", 42.940498, -78.732201}
	case "BUN":
		return Location{"Buenaventura", "America/Bogota", 3.819630, -76.989800}
	case "BUP":
		return Location{"", "Asia/Kolkata", 30.270100, 74.755798}
	case "BUQ":
		return Location{"Bulawayo", "Africa/Harare", -20.017401, 28.617901}
	case "BUR":
		return Location{"Burbank", "America/Los_Angeles", 34.200699, -118.359001}
	case "BUS":
		return Location{"Batumi", "Asia/Tbilisi", 41.610298, 41.599701}
	case "BUW":
		return Location{"Bau Bau-Butung Island", "Asia/Makassar", -5.486880, 122.569000}
	case "BUX":
		return Location{"", "Africa/Lubumbashi", 1.565720, 30.220800}
	case "BUZ":
		return Location{"Bushehr", "Asia/Tehran", 28.944799, 50.834599}
	case "BVA":
		return Location{"Beauvais/Tille", "Europe/Paris", 49.454399, 2.112780}
	case "BVB":
		return Location{"Boa Vista", "America/Boa_Vista", 2.841389, -60.692223}
	case "BVC":
		return Location{"Rabil", "Atlantic/Cape_Verde", 16.136499, -22.888901}
	case "BVE":
		return Location{"Brive-la-Gaillarde", "Europe/Paris", 45.150799, 1.469170}
	case "BVG":
		return Location{"Berlevag", "Europe/Oslo", 70.871399, 29.034201}
	case "BVH":
		return Location{"Vilhena", "America/Cuiaba", -12.694400, -60.098301}
	case "BVI":
		return Location{"", "Australia/Brisbane", -25.897499, 139.348007}
	case "BVS":
		return Location{"Breves", "America/Belem", -1.636530, -50.443600}
	case "BWA":
		return Location{"Bhairawa", "Asia/Kathmandu", 27.505699, 83.416298}
	case "BWI":
		return Location{"Baltimore", "America/New_York", 39.175400, -76.668297}
	case "BWK":
		return Location{"Brac Island", "Europe/Zagreb", 43.285702, 16.679701}
	case "BWN":
		return Location{"Bandar Seri Begawan", "Asia/Brunei", 4.944200, 114.928001}
	case "BWT":
		return Location{"Burnie", "Australia/Hobart", -40.998901, 145.731003}
	case "BXG":
		return Location{"", "Australia/Melbourne", -36.739399, 144.330002}
	case "BXU":
		return Location{"Butuan City", "Asia/Manila", 8.951320, 125.477997}
	case "BXY":
		return Location{"Baikonur", "Asia/Qyzylorda", 45.622002, 63.215000}
	case "BYC":
		return Location{"Yacuiba", "America/La_Paz", -21.960899, -63.651699}
	case "BYK":
		return Location{"", "Africa/Abidjan", 7.738800, -5.073670}
	case "BYN":
		return Location{"Bayankhongor", "Asia/Ulaanbaatar", 46.163300, 100.704002}
	case "BYO":
		return Location{"Bonito", "America/Campo_Grande", -21.229445, -56.456112}
	case "BZE":
		return Location{"Belize City", "America/Belize", 17.539101, -88.308197}
	case "BZG":
		return Location{"Bydgoszcz", "Europe/Warsaw", 53.096802, 17.977699}
	case "BZL":
		return Location{"Barisal", "Asia/Dhaka", 22.801001, 90.301201}
	case "BZN":
		return Location{"Bozeman", "America/Denver", 45.777500, -111.153000}
	case "BZO":
		return Location{"Bolzano", "Europe/Rome", 46.460201, 11.326400}
	case "BZR":
		return Location{"Beziers/Vias", "Europe/Paris", 43.323502, 3.353900}
	case "BZV":
		return Location{"Brazzaville", "Africa/Brazzaville", -4.251700, 15.253000}
	case "CAB":
		return Location{"Cabinda", "Africa/Luanda", -5.596990, 12.188400}
	case "CAC":
		return Location{"Cascavel", "America/Sao_Paulo", -25.000299, -53.500801}
	case "CAE":
		return Location{"Columbia", "America/New_York", 33.938801, -81.119499}
	case "CAG":
		return Location{"Cagliari", "Europe/Rome", 39.251499, 9.054280}
	case "CAH":
		return Location{"Ca Mau City", "Asia/Ho_Chi_Minh", 9.177667, 105.177778}
	case "CAI":
		return Location{"Cairo", "Africa/Cairo", 30.121901, 31.405600}
	case "CAK":
		return Location{"Akron", "America/New_York", 40.916100, -81.442200}
	case "CAL":
		return Location{"Campbeltown", "Europe/London", 55.437199, -5.686390}
	case "CAN":
		return Location{"Guangzhou", "Asia/Shanghai", 23.392401, 113.299004}
	case "CAP":
		return Location{"Cap Haitien", "America/Port-au-Prince", 19.733000, -72.194702}
	case "CAU":
		return Location{"Caruaru", "America/Recife", -8.282390, -36.013500}
	case "CAW":
		return Location{"Campos Dos Goytacazes", "America/Sao_Paulo", -21.698299, -41.301701}
	case "CAY":
		return Location{"Cayenne / Rochambeau", "America/Cayenne", 4.819810, -52.360401}
	case "CAZ":
		return Location{"", "Australia/Sydney", -31.538300, 145.794006}
	case "CBB":
		return Location{"Cochabamba", "America/La_Paz", -17.421101, -66.177101}
	case "CBG":
		return Location{"Cambridge", "Europe/London", 52.205002, 0.175000}
	case "CBH":
		return Location{"Bechar", "Africa/Algiers", 31.645700, -2.269860}
	case "CBL":
		return Location{"", "America/Caracas", 8.122161, -63.536957}
	case "CBO":
		return Location{"Cotabato City", "Asia/Manila", 7.165240, 124.209999}
	case "CBQ":
		return Location{"Calabar", "Africa/Lagos", 4.976020, 8.347200}
	case "CBR":
		return Location{"Canberra", "Australia/Sydney", -35.306900, 149.195007}
	case "CBT":
		return Location{"Catumbela", "Africa/Luanda", -12.479200, 13.486900}
	case "CCC":
		return Location{"Cayo Coco", "America/Havana", 22.461000, -78.328400}
	case "CCF":
		return Location{"Carcassonne/Salvaza", "Europe/Paris", 43.216000, 2.306320}
	case "CCJ":
		return Location{"Calicut", "Asia/Kolkata", 11.136800, 75.955299}
	case "CCK":
		return Location{"Cocos (Keeling) Islands", "Indian/Cocos", -12.188300, 96.833900}
	case "CCP":
		return Location{"Concepcion", "America/Santiago", -36.772701, -73.063103}
	case "CCR":
		return Location{"Concord", "America/Los_Angeles", 37.989700, -122.056999}
	case "CCS":
		return Location{"Caracas", "America/Caracas", 10.603117, -66.990585}
	case "CCU":
		return Location{"Kolkata", "Asia/Kolkata", 22.654699, 88.446701}
	case "CCV":
		return Location{"Craig Cove", "Pacific/Efate", -16.264999, 167.923996}
	case "CDB":
		return Location{"Cold Bay", "America/Nome", 55.206100, -162.725006}
	case "CDC":
		return Location{"Cedar City", "America/Denver", 37.701000, -113.098999}
	case "CDG":
		return Location{"Paris", "Europe/Paris", 49.012798, 2.550000}
	case "CDP":
		return Location{"", "Asia/Kolkata", 14.510000, 78.772797}
	case "CDR":
		return Location{"Chadron", "America/Denver", 42.837601, -103.095001}
	case "CDT":
		return Location{"Calamocha", "Europe/Madrid", 40.900002, -1.304120}
	case "CDV":
		return Location{"Cordova", "America/Anchorage", 60.491798, -145.477997}
	case "CEB":
		return Location{"Lapu-Lapu City", "Asia/Manila", 10.307500, 123.978996}
	case "CEC":
		return Location{"Crescent City", "America/Los_Angeles", 41.780201, -124.236999}
	case "CED":
		return Location{"", "Australia/Adelaide", -32.130600, 133.710007}
	case "CEE":
		return Location{"Cherepovets", "Europe/Moscow", 59.276667, 38.028333}
	case "CEI":
		return Location{"Chiang Rai", "Asia/Bangkok", 19.952299, 99.882896}
	case "CEK":
		return Location{"Chelyabinsk", "Asia/Yekaterinburg", 55.305801, 61.503300}
	case "CEM":
		return Location{"Central", "America/Anchorage", 65.573799, -144.783005}
	case "CEN":
		return Location{"Ciudad Obregon", "America/Hermosillo", 27.392599, -109.833000}
	case "CEZ":
		return Location{"Cortez", "America/Denver", 37.303001, -108.627998}
	case "CFB":
		return Location{"Cabo Frio", "America/Sao_Paulo", -22.921700, -42.074299}
	case "CFE":
		return Location{"Clermont-Ferrand/Auvergne", "Europe/Paris", 45.786701, 3.169170}
	case "CFG":
		return Location{"Cienfuegos", "America/Havana", 22.150000, -80.414200}
	case "CFN":
		return Location{"Donegal", "Europe/Dublin", 55.044201, -8.341000}
	case "CFR":
		return Location{"Caen/Carpiquet", "Europe/Paris", 49.173302, -0.450000}
	case "CFS":
		return Location{"Coffs Harbour", "Australia/Sydney", -30.320601, 153.115997}
	case "CFU":
		return Location{"Kerkyra Island", "Europe/Athens", 39.601898, 19.911699}
	case "CGB":
		return Location{"Cuiaba", "America/Cuiaba", -15.652900, -56.116699}
	case "CGD":
		return Location{"Changde", "Asia/Shanghai", 28.918900, 111.639999}
	case "CGH":
		return Location{"Sao Paulo", "America/Sao_Paulo", -23.626110, -46.656387}
	case "CGI":
		return Location{"Cape Girardeau", "America/Chicago", 37.225300, -89.570801}
	case "CGK":
		return Location{"Jakarta", "Asia/Jakarta", -6.125570, 106.655998}
	case "CGM":
		return Location{"", "Asia/Manila", 9.253520, 124.707001}
	case "CGN":
		return Location{"Cologne", "Europe/Berlin", 50.865898, 7.142740}
	case "CGO":
		return Location{"Zhengzhou", "Asia/Shanghai", 34.519699, 113.841003}
	case "CGP":
		return Location{"Chittagong", "Asia/Dhaka", 22.249599, 91.813301}
	case "CGQ":
		return Location{"Changchun", "Asia/Shanghai", 43.996201, 125.684998}
	case "CGR":
		return Location{"Campo Grande", "America/Campo_Grande", -20.468700, -54.672501}
	case "CGY":
		return Location{"Cagayan De Oro City", "Asia/Manila", 8.141660, 125.116997}
	case "CHA":
		return Location{"Chattanooga", "America/New_York", 35.035301, -85.203796}
	case "CHC":
		return Location{"Christchurch", "Pacific/Auckland", -43.489399, 172.531998}
	case "CHG":
		return Location{"Chaoyang", "Asia/Shanghai", 41.538101, 120.434998}
	case "CHH":
		return Location{"Chachapoyas", "America/Lima", -6.201810, -77.856102}
	case "CHO":
		return Location{"Charlottesville", "America/New_York", 38.138599, -78.452904}
	case "CHQ":
		return Location{"Souda", "Europe/Athens", 35.531700, 24.149700}
	case "CHS":
		return Location{"Charleston", "America/New_York", 32.898602, -80.040497}
	case "CHT":
		return Location{"Waitangi", "Pacific/Chatham", -43.810001, -176.457001}
	case "CHU":
		return Location{"Chuathbaluk", "America/Anchorage", 61.579102, -159.216003}
	case "CHX":
		return Location{"Changuinola", "America/Panama", 9.458640, -82.516800}
	case "CHY":
		return Location{"", "Pacific/Guadalcanal", -6.711944, 156.396111}
	case "CIA":
		return Location{"Roma", "Europe/Rome", 41.799400, 12.594900}
	case "CID":
		return Location{"Cedar Rapids", "America/Chicago", 41.884701, -91.710800}
	case "CIF":
		return Location{"Chifeng", "Asia/Shanghai", 42.235001, 118.907997}
	case "CIH":
		return Location{"Changzhi", "Asia/Shanghai", 36.247501, 113.125999}
	case "CIJ":
		return Location{"Cobija", "America/La_Paz", -11.040400, -68.782997}
	case "CIK":
		return Location{"Chalkyitsik", "America/Anchorage", 66.644997, -143.740005}
	case "CIT":
		return Location{"Shymkent", "Asia/Almaty", 42.364201, 69.478897}
	case "CIU":
		return Location{"Sault Ste Marie", "America/Detroit", 46.250801, -84.472397}
	case "CIX":
		return Location{"Chiclayo", "America/Lima", -6.787480, -79.828102}
	case "CIY":
		return Location{"Comiso", "Europe/Rome", 36.994601, 14.607182}
	case "CIZ":
		return Location{"Coari", "America/Manaus", -4.134060, -63.132599}
	case "CJA":
		return Location{"Cajamarca", "America/Lima", -7.139180, -78.489403}
	case "CJB":
		return Location{"Coimbatore", "Asia/Kolkata", 11.030000, 77.043404}
	case "CJC":
		return Location{"Calama", "America/Santiago", -22.498199, -68.903603}
	case "CJJ":
		return Location{"Cheongju", "Asia/Seoul", 36.716599, 127.499001}
	case "CJM":
		return Location{"", "Asia/Bangkok", 10.711200, 99.361702}
	case "CJS":
		return Location{"Ciudad Juarez", "America/Ojinaga", 31.636101, -106.429001}
	case "CJU":
		return Location{"Jeju City", "Asia/Seoul", 33.511299, 126.492996}
	case "CKB":
		return Location{"Clarksburg", "America/New_York", 39.296600, -80.228104}
	case "CKG":
		return Location{"Chongqing", "Asia/Shanghai", 29.719200, 106.641998}
	case "CKH":
		return Location{"Chokurdah", "Asia/Srednekolymsk", 70.623100, 147.901993}
	case "CKO":
		return Location{"Cornelio Procopio", "America/Sao_Paulo", -23.152500, -50.602501}
	case "CKS":
		return Location{"Carajas", "America/Belem", -6.115278, -50.001389}
	case "CKY":
		return Location{"Conakry", "Africa/Conakry", 9.576890, -13.612000}
	case "CKZ":
		return Location{"Canakkale", "Europe/Istanbul", 40.137699, 26.426800}
	case "CLD":
		return Location{"Carlsbad", "America/Los_Angeles", 33.128300, -117.279999}
	case "CLE":
		return Location{"Cleveland", "America/New_York", 41.411701, -81.849800}
	case "CLJ":
		return Location{"Cluj-Napoca", "Europe/Bucharest", 46.785198, 23.686199}
	case "CLL":
		return Location{"College Station", "America/Chicago", 30.588600, -96.363800}
	case "CLO":
		return Location{"Cali", "America/Bogota", 3.543220, -76.381600}
	case "CLP":
		return Location{"Clarks Point", "America/Anchorage", 58.833698, -158.529007}
	case "CLQ":
		return Location{"Colima", "America/Mexico_City", 19.277000, -103.577003}
	case "CLT":
		return Location{"Charlotte", "America/New_York", 35.214001, -80.943100}
	case "CLV":
		return Location{"Caldas Novas", "America/Sao_Paulo", -17.725300, -48.607498}
	case "CLY":
		return Location{"Calvi/Sainte-Catherine", "Europe/Paris", 42.530800, 8.793190}
	case "CMA":
		return Location{"", "Australia/Brisbane", -28.030001, 145.621994}
	case "CMB":
		return Location{"Colombo", "Asia/Colombo", 7.180760, 79.884102}
	case "CME":
		return Location{"Ciudad del Carmen", "America/Merida", 18.653700, -91.799004}
	case "CMF":
		return Location{"Chambery/Aix-les-Bains", "Europe/Paris", 45.638100, 5.880230}
	case "CMG":
		return Location{"Corumba", "America/Campo_Grande", -19.011944, -57.671391}
	case "CMH":
		return Location{"Columbus", "America/New_York", 39.998001, -82.891899}
	case "CMI":
		return Location{"Champaign/Urbana", "America/Chicago", 40.039200, -88.278099}
	case "CMN":
		return Location{"Casablanca", "Africa/Casablanca", 33.367500, -7.589970}
	case "CMW":
		return Location{"Camaguey", "America/Havana", 21.420300, -77.847504}
	case "CMX":
		return Location{"Hancock", "America/Detroit", 47.168400, -88.489098}
	case "CNC":
		return Location{"", "Australia/Brisbane", -10.050000, 143.070007}
	case "CND":
		return Location{"Constanta", "Europe/Bucharest", 44.362202, 28.488300}
	case "CNF":
		return Location{"Belo Horizonte", "America/Sao_Paulo", -19.624443, -43.971943}
	case "CNJ":
		return Location{"Cloncurry", "Australia/Brisbane", -20.668600, 140.503998}
	case "CNM":
		return Location{"Carlsbad", "America/Denver", 32.337502, -104.263000}
	case "CNN":
		return Location{"Chulman", "Asia/Yakutsk", 56.913898, 124.914001}
	case "CNQ":
		return Location{"Corrientes", "America/Argentina/Cordoba", -27.445500, -58.761900}
	case "CNS":
		return Location{"Cairns", "Australia/Brisbane", -16.885799, 145.755005}
	case "CNX":
		return Location{"Chiang Mai", "Asia/Bangkok", 18.766800, 98.962601}
	case "CNY":
		return Location{"Moab", "America/Denver", 38.755001, -109.754997}
	case "COD":
		return Location{"Cody", "America/Denver", 44.520199, -109.024002}
	case "COH":
		return Location{"", "Asia/Kolkata", 26.330500, 89.467201}
	case "COK":
		return Location{"Cochin", "Asia/Kolkata", 10.152000, 76.401901}
	case "COO":
		return Location{"Cotonou", "Africa/Porto-Novo", 6.357230, 2.384350}
	case "COR":
		return Location{"Cordoba", "America/Argentina/Cordoba", -31.323601, -64.208000}
	case "COS":
		return Location{"Colorado Springs", "America/Denver", 38.805801, -104.700996}
	case "COU":
		return Location{"Columbia", "America/Chicago", 38.818100, -92.219597}
	case "CPC":
		return Location{"Chapelco/San Martin de los Andes", "America/Argentina/Salta", -40.075401, -71.137299}
	case "CPD":
		return Location{"", "Australia/Adelaide", -29.040001, 134.720993}
	case "CPE":
		return Location{"Campeche", "America/Merida", 19.816799, -90.500298}
	case "CPH":
		return Location{"Copenhagen", "Europe/Copenhagen", 55.617901, 12.656000}
	case "CPO":
		return Location{"Copiapo", "America/Santiago", -27.261200, -70.779198}
	case "CPR":
		return Location{"Casper", "America/Denver", 42.908001, -106.463997}
	case "CPT":
		return Location{"Cape Town", "Africa/Johannesburg", -33.964802, 18.601700}
	case "CPV":
		return Location{"Campina Grande", "America/Fortaleza", -7.269920, -35.896400}
	case "CPX":
		return Location{"Culebra Island", "America/Puerto_Rico", 18.313289, -65.304324}
	case "CQD":
		return Location{"Shahrekord", "Asia/Tehran", 32.297199, 50.842201}
	case "CRA":
		return Location{"Craiova", "Europe/Bucharest", 44.318100, 23.888599}
	case "CRD":
		return Location{"Comodoro Rivadavia", "America/Argentina/Catamarca", -45.785300, -67.465500}
	case "CRI":
		return Location{"Colonel Hill", "America/Nassau", 22.745600, -74.182404}
	case "CRK":
		return Location{"Angeles City", "Asia/Manila", 15.186000, 120.559998}
	case "CRL":
		return Location{"Brussels", "Europe/Brussels", 50.459202, 4.453820}
	case "CRM":
		return Location{"Catarman", "Asia/Manila", 12.502400, 124.636002}
	case "CRP":
		return Location{"Corpus Christi", "America/Chicago", 27.770399, -97.501198}
	case "CRV":
		return Location{"Crotone", "Europe/Rome", 38.997200, 17.080200}
	case "CRW":
		return Location{"Charleston", "America/New_York", 38.373100, -81.593201}
	case "CRZ":
		return Location{"Turkmenabat", "Asia/Ashgabat", 39.083302, 63.613300}
	case "CSG":
		return Location{"Columbus", "America/New_York", 32.516300, -84.938904}
	case "CSK":
		return Location{"Cap Skirring", "Africa/Dakar", 12.410200, -16.746099}
	case "CSU":
		return Location{"Santa Cruz Do Sul", "America/Sao_Paulo", -29.684099, -52.412201}
	case "CSX":
		return Location{"Changsha", "Asia/Shanghai", 28.189199, 113.220001}
	case "CSY":
		return Location{"Cheboksary", "Europe/Moscow", 56.090302, 47.347301}
	case "CTA":
		return Location{"Catania", "Europe/Rome", 37.466801, 15.066400}
	case "CTC":
		return Location{"Catamarca", "America/Argentina/Catamarca", -28.595600, -65.751701}
	case "CTD":
		return Location{"Chitre", "America/Panama", 7.987840, -80.409698}
	case "CTG":
		return Location{"Cartagena", "America/Bogota", 10.442400, -75.513000}
	case "CTL":
		return Location{"Charleville", "Australia/Brisbane", -26.413300, 146.261993}
	case "CTM":
		return Location{"Chetumal", "America/Cancun", 18.504700, -88.326797}
	case "CTS":
		return Location{"Chitose / Tomakomai", "Asia/Tokyo", 42.775200, 141.692001}
	case "CTU":
		return Location{"Chengdu", "Asia/Shanghai", 30.578501, 103.946999}
	case "CUC":
		return Location{"Cucuta", "America/Bogota", 7.927570, -72.511500}
	case "CUE":
		return Location{"Cuenca", "America/Guayaquil", -2.889470, -78.984398}
	case "CUF":
		return Location{"Cuneo", "Europe/Rome", 44.547001, 7.623220}
	case "CUL":
		return Location{"Culiacan", "America/Mazatlan", 24.764500, -107.474998}
	case "CUM":
		return Location{"", "America/Caracas", 10.450333, -64.130470}
	case "CUN":
		return Location{"Cancun", "America/Cancun", 21.036501, -86.877098}
	case "CUR":
		return Location{"Willemstad", "America/Curacao", 12.188900, -68.959801}
	case "CUU":
		return Location{"Chihuahua", "America/Chihuahua", 28.702900, -105.964996}
	case "CUZ":
		return Location{"Cusco", "America/Lima", -13.535700, -71.938797}
	case "CVG":
		return Location{"Hebron", "America/New_York", 39.048801, -84.667801}
	case "CVM":
		return Location{"Ciudad Victoria", "America/Monterrey", 23.703300, -98.956497}
	case "CVN":
		return Location{"Clovis", "America/Denver", 34.425098, -103.079002}
	case "CVQ":
		return Location{"", "Australia/Perth", -24.880600, 113.671997}
	case "CVU":
		return Location{"Corvo", "Atlantic/Azores", 39.671501, -31.113600}
	case "CWA":
		return Location{"Mosinee", "America/Chicago", 44.777599, -89.666801}
	case "CWB":
		return Location{"Curitiba", "America/Sao_Paulo", -25.528500, -49.175800}
	case "CWL":
		return Location{"Cardiff", "Europe/London", 51.396702, -3.343330}
	case "CXB":
		return Location{"Cox's Bazar", "Asia/Dhaka", 21.452200, 91.963898}
	case "CXI":
		return Location{"Banana", "Pacific/Kiritimati", 1.986160, -157.350006}
	case "CXJ":
		return Location{"Caxias Do Sul", "America/Sao_Paulo", -29.197100, -51.187500}
	case "CXR":
		return Location{"Nha Trang", "Asia/Ho_Chi_Minh", 11.998200, 109.219002}
	case "CXY":
		return Location{"Cat Cay", "America/Nassau", 25.600000, -79.266998}
	case "CYA":
		return Location{"Les Cayes", "America/Port-au-Prince", 18.271099, -73.788300}
	case "CYB":
		return Location{"Cayman Brac", "America/Cayman", 19.687000, -79.882797}
	case "CYF":
		return Location{"Chefornak", "America/Nome", 60.149200, -164.285995}
	case "CYI":
		return Location{"Chiayi City", "Asia/Taipei", 23.461800, 120.392998}
	case "CYO":
		return Location{"Cayo Largo del Sur", "America/Havana", 21.616501, -81.545998}
	case "CYP":
		return Location{"Calbayog City", "Asia/Manila", 12.072700, 124.544998}
	case "CYS":
		return Location{"Cheyenne", "America/Denver", 41.155701, -104.811997}
	case "CYX":
		return Location{"Cherskiy", "Asia/Srednekolymsk", 68.740601, 161.337997}
	case "CYZ":
		return Location{"Cauayan City", "Asia/Manila", 16.929899, 121.752998}
	case "CZL":
		return Location{"Constantine", "Africa/Algiers", 36.276001, 6.620390}
	case "CZM":
		return Location{"Cozumel", "America/Cancun", 20.522400, -86.925598}
	case "CZS":
		return Location{"Cruzeiro Do Sul", "America/Rio_Branco", -7.599910, -72.769501}
	case "CZU":
		return Location{"Corozal", "America/Bogota", 9.332740, -75.285600}
	case "CZX":
		return Location{"Changzhou", "Asia/Shanghai", 31.919701, 119.778999}
	case "DAB":
		return Location{"Daytona Beach", "America/New_York", 29.179899, -81.058098}
	case "DAC":
		return Location{"Dhaka", "Asia/Dhaka", 23.843347, 90.397783}
	case "DAD":
		return Location{"Da Nang", "Asia/Ho_Chi_Minh", 16.043900, 108.198997}
	case "DAL":
		return Location{"Dallas", "America/Chicago", 32.847099, -96.851799}
	case "DAM":
		return Location{"Damascus", "Asia/Damascus", 33.411499, 36.515598}
	case "DAR":
		return Location{"Dar es Salaam", "Africa/Dar_es_Salaam", -6.878110, 39.202599}
	case "DAT":
		return Location{"Datong", "Asia/Shanghai", 40.060299, 113.482002}
	case "DAU":
		return Location{"Daru", "Pacific/Port_Moresby", -9.086760, 143.207993}
	case "DAV":
		return Location{"David", "America/Panama", 8.391000, -82.434998}
	case "DAY":
		return Location{"Dayton", "America/New_York", 39.902401, -84.219398}
	case "DBO":
		return Location{"Dubbo", "Australia/Sydney", -32.216702, 148.574997}
	case "DBQ":
		return Location{"Dubuque", "America/Chicago", 42.402000, -90.709503}
	case "DBR":
		return Location{"Darbhanga", "Asia/Kolkata", 26.194722, 85.917500}
	case "DBV":
		return Location{"Dubrovnik", "Europe/Zagreb", 42.561401, 18.268200}
	case "DCA":
		return Location{"Washington", "America/New_York", 38.852100, -77.037697}
	case "DCM":
		return Location{"Castres/Mazamet", "Europe/Paris", 43.556301, 2.289180}
	case "DDC":
		return Location{"Dodge City", "America/Chicago", 37.763401, -99.965599}
	case "DDG":
		return Location{"Dandong", "Asia/Shanghai", 40.024700, 124.286003}
	case "DEB":
		return Location{"Debrecen", "Europe/Budapest", 47.488899, 21.615299}
	case "DEC":
		return Location{"Decatur", "America/Chicago", 39.834599, -88.865700}
	case "DED":
		return Location{"Dehradun", "Asia/Kolkata", 30.189699, 78.180298}
	case "DEE":
		return Location{"Kunashir Island", "Asia/Ust-Nera", 43.958401, 145.682999}
	case "DEL":
		return Location{"New Delhi", "Asia/Kolkata", 28.566500, 77.103104}
	case "DEN":
		return Location{"Denver", "America/Denver", 39.861698, -104.672997}
	case "DFW":
		return Location{"Dallas-Fort Worth", "America/Chicago", 32.896801, -97.038002}
	case "DGE":
		return Location{"Mudgee", "Australia/Sydney", -32.562500, 149.610992}
	case "DGH":
		return Location{"Deoghar", "Asia/Kolkata", 24.444722, 86.702500}
	case "DGO":
		return Location{"Durango", "America/Monterrey", 24.124201, -104.528000}
	case "DGT":
		return Location{"Dumaguete City", "Asia/Manila", 9.333710, 123.300003}
	case "DHI":
		return Location{"Dhangarhi", "Asia/Kathmandu", 28.753300, 80.581902}
	case "DHM":
		return Location{"", "Asia/Kolkata", 32.165100, 76.263397}
	case "DHN":
		return Location{"Dothan", "America/Chicago", 31.321301, -85.449600}
	case "DIB":
		return Location{"Dibrugarh", "Asia/Kolkata", 27.483900, 95.016899}
	case "DIE":
		return Location{"", "Indian/Antananarivo", -12.349400, 49.291698}
	case "DIG":
		return Location{"Shangri-La", "Asia/Shanghai", 27.793600, 99.677200}
	case "DIK":
		return Location{"Dickinson", "America/Denver", 46.797401, -102.802002}
	case "DIL":
		return Location{"Dili", "Asia/Dili", -8.546400, 125.526001}
	case "DIN":
		return Location{"Dien Bien Phu", "Asia/Bangkok", 21.397499, 103.008003}
	case "DIR":
		return Location{"Dire Dawa", "Africa/Addis_Ababa", 9.624700, 41.854198}
	case "DIU":
		return Location{"Diu", "Asia/Kolkata", 20.713100, 70.921097}
	case "DIY":
		return Location{"Diyarbakir", "Europe/Istanbul", 37.893902, 40.201000}
	case "DJB":
		return Location{"Jambi-Sumatra Island", "Asia/Jakarta", -1.638020, 103.643997}
	case "DJE":
		return Location{"Djerba", "Africa/Tunis", 33.875000, 10.775500}
	case "DJG":
		return Location{"Djanet", "Africa/Algiers", 24.292801, 9.452440}
	case "DJJ":
		return Location{"Jayapura-Papua Island", "Asia/Jayapura", -2.576950, 140.516006}
	case "DKR":
		return Location{"Dakar", "Africa/Dakar", 14.739700, -17.490200}
	case "DKS":
		return Location{"Dikson", "Asia/Krasnoyarsk", 73.517807, 80.379669}
	case "DLA":
		return Location{"Douala", "Africa/Douala", 4.006080, 9.719480}
	case "DLC":
		return Location{"Dalian", "Asia/Shanghai", 38.965698, 121.539001}
	case "DLE":
		return Location{"Dole/Tavaux", "Europe/Paris", 47.039001, 5.427250}
	case "DLG":
		return Location{"Dillingham", "America/Anchorage", 59.044701, -158.505005}
	case "DLH":
		return Location{"Duluth", "America/Chicago", 46.842098, -92.193604}
	case "DLI":
		return Location{"Dalat", "Asia/Ho_Chi_Minh", 11.750000, 108.366997}
	case "DLM":
		return Location{"Dalaman", "Europe/Istanbul", 36.713100, 28.792500}
	case "DLU":
		return Location{"Xiaguan", "Asia/Shanghai", 25.649401, 100.319000}
	case "DLY":
		return Location{"Dillon's Bay", "Pacific/Efate", -18.769400, 169.001007}
	case "DLZ":
		return Location{"Dalanzadgad", "Asia/Ulaanbaatar", 43.591702, 104.430000}
	case "DMB":
		return Location{"Taraz", "Asia/Almaty", 42.853600, 71.303596}
	case "DMD":
		return Location{"", "Australia/Brisbane", -17.940300, 138.822006}
	case "DME":
		return Location{"Moscow", "Europe/Moscow", 55.408798, 37.906300}
	case "DMK":
		return Location{"Bangkok", "Asia/Bangkok", 13.912600, 100.607002}
	case "DMM":
		return Location{"Ad Dammam", "Asia/Riyadh", 26.471201, 49.797901}
	case "DMU":
		return Location{"Dimapur", "Asia/Kolkata", 25.883900, 93.771103}
	case "DND":
		return Location{"Dundee", "Europe/London", 56.452499, -3.025830}
	case "DNH":
		return Location{"Dunhuang", "Asia/Shanghai", 40.161098, 94.809196}
	case "DNZ":
		return Location{"Denizli", "Europe/Istanbul", 37.785599, 29.701300}
	case "DOB":
		return Location{"Dobo-Kobror Island", "Asia/Jayapura", -5.772220, 134.212006}
	case "DOD":
		return Location{"Dodoma", "Africa/Dar_es_Salaam", -6.170440, 35.752602}
	case "DOH":
		return Location{"Doha", "Asia/Qatar", 25.260595, 51.613766}
	case "DOL":
		return Location{"Deauville", "Europe/Paris", 49.365299, 0.154306}
	case "DOM":
		return Location{"Marigot", "America/Dominica", 15.547000, -61.299999}
	case "DOY":
		return Location{"Dongying", "Asia/Shanghai", 37.508598, 118.788002}
	case "DPL":
		return Location{"Dipolog City", "Asia/Manila", 8.601983, 123.341875}
	case "DPO":
		return Location{"Devonport", "Australia/Hobart", -41.169701, 146.429993}
	case "DPS":
		return Location{"Denpasar-Bali Island", "Asia/Makassar", -8.748170, 115.167000}
	case "DQA":
		return Location{"Daqing Shi", "Asia/Shanghai", 46.746389, 125.140556}
	case "DQM":
		return Location{"Duqm", "Asia/Muscat", 19.501900, 57.634200}
	case "DRB":
		return Location{"", "Australia/Perth", -17.370001, 123.661003}
	case "DRG":
		return Location{"Deering", "America/Nome", 66.069603, -162.766006}
	case "DRK":
		return Location{"Puntarenas", "America/Costa_Rica", 8.718890, -83.641701}
	case "DRO":
		return Location{"Durango", "America/Denver", 37.151501, -107.753998}
	case "DRS":
		return Location{"Dresden", "Europe/Berlin", 51.132801, 13.767200}
	case "DRW":
		return Location{"Darwin", "Australia/Darwin", -12.414700, 130.876999}
	case "DSE":
		return Location{"Dessie", "Africa/Addis_Ababa", 11.082500, 39.711399}
	case "DSM":
		return Location{"Des Moines", "America/Chicago", 41.534000, -93.663101}
	case "DSN":
		return Location{"Ordos", "Asia/Shanghai", 39.490000, 109.861389}
	case "DSS":
		return Location{"Diass", "Africa/Dakar", 14.671111, -17.066944}
	case "DTB":
		return Location{"Siborong-Borong", "Asia/Jakarta", 2.259722, 98.995278}
	case "DTM":
		return Location{"Dortmund", "Europe/Berlin", 51.518299, 7.612240}
	case "DTW":
		return Location{"Detroit", "America/Detroit", 42.212399, -83.353401}
	case "DUB":
		return Location{"Dublin", "Europe/Dublin", 53.421299, -6.270070}
	case "DUD":
		return Location{"Dunedin", "Pacific/Auckland", -45.928101, 170.197998}
	case "DUE":
		return Location{"Chitato", "Africa/Luanda", -7.400890, 20.818501}
	case "DUJ":
		return Location{"Dubois", "America/New_York", 41.178299, -78.898697}
	case "DUR":
		return Location{"Durban", "Africa/Johannesburg", -29.614444, 31.119722}
	case "DUS":
		return Location{"Dusseldorf", "Europe/Berlin", 51.289501, 6.766780}
	case "DUT":
		return Location{"Unalaska", "America/Nome", 53.900101, -166.544006}
	case "DVL":
		return Location{"Devils Lake", "America/Chicago", 48.114201, -98.908798}
	case "DVO":
		return Location{"Davao City", "Asia/Manila", 7.125520, 125.646004}
	case "DWC":
		return Location{"Jebel Ali", "Asia/Dubai", 24.896667, 55.161389}
	case "DWD":
		return Location{"Dawadmi", "Asia/Riyadh", 24.500000, 44.400002}
	case "DXB":
		return Location{"Dubai", "Asia/Dubai", 25.252800, 55.364399}
	case "DYG":
		return Location{"Dayong", "Asia/Shanghai", 29.102800, 110.443001}
	case "DYR":
		return Location{"Anadyr", "Asia/Anadyr", 64.734901, 177.740997}
	case "DYU":
		return Location{"Dushanbe", "Asia/Dushanbe", 38.543301, 68.824997}
	case "DZA":
		return Location{"Dzaoudzi", "Indian/Mayotte", -12.804700, 45.281101}
	case "DZN":
		return Location{"Zhezkazgan", "Asia/Almaty", 47.708302, 67.733299}
	case "EAA":
		return Location{"Eagle", "America/Anchorage", 64.776398, -141.151001}
	case "EAE":
		return Location{"Sangafa", "Pacific/Efate", -17.090300, 168.343002}
	case "EAM":
		return Location{"", "Asia/Riyadh", 17.611401, 44.419201}
	case "EAR":
		return Location{"Kearney", "America/Chicago", 40.727001, -99.006798}
	case "EAS":
		return Location{"Hondarribia", "Europe/Madrid", 43.356499, -1.790610}
	case "EAT":
		return Location{"Wenatchee", "America/Los_Angeles", 47.398899, -120.207001}
	case "EAU":
		return Location{"Eau Claire", "America/Chicago", 44.865799, -91.484299}
	case "EBB":
		return Location{"Kampala", "Africa/Kampala", 0.042386, 32.443501}
	case "EBH":
		return Location{"El Bayadh", "Africa/Algiers", 33.721667, 1.092500}
	case "EBJ":
		return Location{"Esbjerg", "Europe/Copenhagen", 55.525902, 8.553400}
	case "EBL":
		return Location{"Arbil", "Asia/Baghdad", 36.237598, 43.963200}
	case "ECN":
		return Location{"Nicosia", "Asia/Famagusta", 35.154701, 33.496101}
	case "ECP":
		return Location{"Panama City Beach", "America/Chicago", 30.341700, -85.797300}
	case "EDI":
		return Location{"Edinburgh", "Europe/London", 55.950001, -3.372500}
	case "EDL":
		return Location{"Eldoret", "Africa/Nairobi", 0.404458, 35.238899}
	case "EDO":
		return Location{"Edremit", "Europe/Istanbul", 39.554600, 27.013800}
	case "EDR":
		return Location{"", "Australia/Brisbane", -14.896700, 141.608994}
	case "EEK":
		return Location{"Eek", "America/Nome", 60.213673, -162.043884}
	case "EFL":
		return Location{"Kefallinia Island", "Europe/Athens", 38.120098, 20.500500}
	case "EGC":
		return Location{"Bergerac/Roumaniere", "Europe/Paris", 44.825298, 0.518611}
	case "EGE":
		return Location{"Eagle", "America/Denver", 39.642601, -106.917999}
	case "EGM":
		return Location{"Sege", "Pacific/Guadalcanal", -8.578890, 157.876007}
	case "EGS":
		return Location{"Egilsstadir", "Atlantic/Reykjavik", 65.283302, -14.401400}
	case "EGX":
		return Location{"Egegik", "America/Anchorage", 58.185501, -157.375000}
	case "EIN":
		return Location{"Eindhoven", "Europe/Amsterdam", 51.450100, 5.374530}
	case "EIS":
		return Location{"Road Town", "America/Tortola", 18.444799, -64.542999}
	case "EJA":
		return Location{"Barrancabermeja", "America/Bogota", 7.024330, -73.806800}
	case "EKO":
		return Location{"Elko", "America/Los_Angeles", 40.824902, -115.792000}
	case "ELC":
		return Location{"Elcho Island", "Australia/Darwin", -12.019400, 135.570999}
	case "ELD":
		return Location{"El Dorado", "America/Chicago", 33.221001, -92.813301}
	case "ELG":
		return Location{"", "Africa/Algiers", 30.571301, 2.859590}
	case "ELH":
		return Location{"North Eleuthera", "America/Nassau", 25.474899, -76.683502}
	case "ELI":
		return Location{"Elim", "America/Nome", 64.614700, -162.272003}
	case "ELM":
		return Location{"Elmira/Corning", "America/New_York", 42.159901, -76.891602}
	case "ELP":
		return Location{"El Paso", "America/Denver", 31.807199, -106.377998}
	case "ELQ":
		return Location{"", "Asia/Riyadh", 26.302799, 43.774399}
	case "ELS":
		return Location{"East London", "Africa/Johannesburg", -33.035599, 27.825899}
	case "ELU":
		return Location{"Guemar", "Africa/Algiers", 33.511398, 6.776790}
	case "EMA":
		return Location{"Nottingham", "Europe/London", 52.831100, -1.328060}
	case "EMD":
		return Location{"Emerald", "Australia/Brisbane", -23.567499, 148.179001}
	case "EMK":
		return Location{"Emmonak", "America/Nome", 62.786098, -164.490997}
	case "EMN":
		return Location{"Nema", "Africa/Nouakchott", 16.622000, -7.316600}
	case "ENA":
		return Location{"Kenai", "America/Anchorage", 60.573101, -151.244995}
	case "ENE":
		return Location{"Ende-Flores Island", "Asia/Makassar", -8.849290, 121.661003}
	case "ENH":
		return Location{"Enshi", "Asia/Shanghai", 30.320299, 109.485001}
	case "ENI":
		return Location{"El Nido", "Asia/Manila", 11.202500, 119.416000}
	case "ENU":
		return Location{"Enegu", "Africa/Lagos", 6.474270, 7.561960}
	case "ENY":
		return Location{"Yan'an", "Asia/Shanghai", 36.636902, 109.554001}
	case "EOH":
		return Location{"Medellin", "America/Bogota", 6.220549, -75.590582}
	case "EPR":
		return Location{"", "Australia/Perth", -33.684399, 121.822998}
	case "EQS":
		return Location{"Esquel", "America/Argentina/Catamarca", -42.908001, -71.139503}
	case "ERC":
		return Location{"Erzincan", "Europe/Istanbul", 39.710201, 39.527000}
	case "ERF":
		return Location{"Erfurt", "Europe/Berlin", 50.979801, 10.958100}
	case "ERG":
		return Location{"Erbogachen", "Asia/Irkutsk", 61.275002, 108.029999}
	case "ERH":
		return Location{"Errachidia", "Africa/Casablanca", 31.947500, -4.398330}
	case "ERI":
		return Location{"Erie", "America/New_York", 42.083127, -80.173867}
	case "ERL":
		return Location{"Erenhot", "Asia/Shanghai", 43.422500, 112.096667}
	case "ERN":
		return Location{"Eirunepe", "America/Eirunepe", -6.639530, -69.879799}
	case "ERS":
		return Location{"Windhoek", "Africa/Windhoek", -22.612200, 17.080400}
	case "ERZ":
		return Location{"Erzurum", "Europe/Istanbul", 39.956501, 41.170200}
	case "ESB":
		return Location{"Ankara", "Europe/Istanbul", 40.128101, 32.995098}
	case "ESC":
		return Location{"Escanaba", "America/Detroit", 45.722698, -87.093697}
	case "ESD":
		return Location{"Eastsound", "America/Los_Angeles", 48.708199, -122.910004}
	case "ESU":
		return Location{"Essaouira", "Africa/Casablanca", 31.397499, -9.681670}
	case "ETM":
		return Location{"Eilat", "Asia/Jerusalem", 29.723694, 35.011417}
	case "ETR":
		return Location{"Santa Rosa", "America/Guayaquil", -3.435160, -79.977798}
	case "ETZ":
		return Location{"Metz / Nancy", "Europe/Paris", 48.982101, 6.251320}
	case "EUA":
		return Location{"Eua Island", "Pacific/Tongatapu", -21.378300, -174.957993}
	case "EUG":
		return Location{"Eugene", "America/Los_Angeles", 44.124599, -123.211998}
	case "EUN":
		return Location{"El Aaiun", "Africa/El_Aaiun", 27.151699, -13.219200}
	case "EUQ":
		return Location{"San Jose", "Asia/Manila", 10.766000, 121.932999}
	case "EUX":
		return Location{"Sint Eustatius", "America/Kralendijk", 17.496500, -62.979401}
	case "EVE":
		return Location{"Evenes", "Europe/Oslo", 68.491302, 16.678101}
	case "EVG":
		return Location{"", "Europe/Stockholm", 62.047798, 14.422900}
	case "EVN":
		return Location{"Yerevan", "Asia/Yerevan", 40.147301, 44.395901}
	case "EVV":
		return Location{"Evansville", "America/Chicago", 38.036999, -87.532402}
	case "EWB":
		return Location{"New Bedford", "America/New_York", 41.676102, -70.956902}
	case "EWN":
		return Location{"New Bern", "America/New_York", 35.073002, -77.042900}
	case "EWR":
		return Location{"Newark", "America/New_York", 40.692501, -74.168701}
	case "EXT":
		return Location{"Exeter", "Europe/London", 50.734402, -3.413890}
	case "EYP":
		return Location{"El Yopal", "America/Bogota", 5.319110, -72.384000}
	case "EYW":
		return Location{"Key West", "America/New_York", 24.556101, -81.759598}
	case "EZE":
		return Location{"Ezeiza", "America/Argentina/Buenos_Aires", -34.822200, -58.535800}
	case "EZS":
		return Location{"Elazig", "Europe/Istanbul", 38.606899, 39.291401}
	case "EZV":
		return Location{"", "Asia/Yekaterinburg", 63.921001, 65.030502}
	case "FAC":
		return Location{"", "Pacific/Tahiti", -16.686701, -145.328995}
	case "FAE":
		return Location{"Vagar", "Atlantic/Faroe", 62.063599, -7.277220}
	case "FAI":
		return Location{"Fairbanks", "America/Anchorage", 64.815102, -147.856003}
	case "FAO":
		return Location{"Faro", "Europe/Lisbon", 37.014400, -7.965910}
	case "FAR":
		return Location{"Fargo", "America/Chicago", 46.920700, -96.815804}
	case "FAT":
		return Location{"Fresno", "America/Los_Angeles", 36.776199, -119.718002}
	case "FAV":
		return Location{"", "Pacific/Tahiti", -16.054100, -145.656998}
	case "FAY":
		return Location{"Fayetteville", "America/New_York", 34.991199, -78.880302}
	case "FBE":
		return Location{"Francisco Beltrao", "America/Sao_Paulo", -26.059200, -53.063499}
	case "FBM":
		return Location{"Lubumbashi", "Africa/Lubumbashi", -11.591300, 27.530899}
	case "FCA":
		return Location{"Kalispell", "America/Denver", 48.310501, -114.255997}
	case "FCO":
		return Location{"Rome", "Europe/Rome", 41.804501, 12.250800}
	case "FDE":
		return Location{"Forde", "Europe/Oslo", 61.391102, 5.756940}
	case "FDF":
		return Location{"Fort-de-France", "America/Martinique", 14.591000, -61.003201}
	case "FDH":
		return Location{"Friedrichshafen", "Europe/Berlin", 47.671299, 9.511490}
	case "FEC":
		return Location{"Feira De Santana", "America/Bahia", -12.200300, -38.906799}
	case "FEG":
		return Location{"Fergana", "Asia/Tashkent", 40.358799, 71.745003}
	case "FEN":
		return Location{"Fernando De Noronha", "America/Noronha", -3.854930, -32.423302}
	case "FEZ":
		return Location{"Fes", "Africa/Casablanca", 33.927299, -4.977960}
	case "FGU":
		return Location{"", "Pacific/Tahiti", -15.819900, -140.886993}
	case "FHZ":
		return Location{"Fakahina", "Pacific/Tahiti", -15.992200, -140.164993}
	case "FIH":
		return Location{"Kinshasa", "Africa/Kinshasa", -4.385750, 15.444600}
	case "FIZ":
		return Location{"", "Australia/Perth", -18.181900, 125.558998}
	case "FJR":
		return Location{"", "Asia/Dubai", 25.112200, 56.324001}
	case "FKB":
		return Location{"Baden-Baden", "Europe/Berlin", 48.779400, 8.080500}
	case "FKI":
		return Location{"Kisangani", "Africa/Lubumbashi", 0.481639, 25.337999}
	case "FKQ":
		return Location{"Fakfak-Papua Island", "Asia/Jayapura", -2.920190, 132.266998}
	case "FKS":
		return Location{"Sukagawa", "Asia/Tokyo", 37.227402, 140.431000}
	case "FLA":
		return Location{"Florencia", "America/Bogota", 1.589190, -75.564400}
	case "FLG":
		return Location{"Flagstaff", "America/Phoenix", 35.138500, -111.670998}
	case "FLL":
		return Location{"Fort Lauderdale", "America/New_York", 26.072599, -80.152702}
	case "FLN":
		return Location{"Florianopolis", "America/Sao_Paulo", -27.670279, -48.552502}
	case "FLO":
		return Location{"Florence", "America/New_York", 34.185398, -79.723900}
	case "FLR":
		return Location{"Firenze", "Europe/Rome", 43.810001, 11.205100}
	case "FLS":
		return Location{"", "Australia/Hobart", -40.091702, 147.992996}
	case "FLW":
		return Location{"Santa Cruz das Flores", "Atlantic/Azores", 39.455299, -31.131399}
	case "FLZ":
		return Location{"Sibolga-Sumatra Island", "Asia/Jakarta", 1.555940, 98.888901}
	case "FMA":
		return Location{"Formosa", "America/Argentina/Cordoba", -26.212700, -58.228100}
	case "FMI":
		return Location{"", "Africa/Lubumbashi", -5.875560, 29.250000}
	case "FMM":
		return Location{"Memmingen", "Europe/Berlin", 47.988800, 10.239500}
	case "FMO":
		return Location{"Munster", "Europe/Berlin", 52.134602, 7.684830}
	case "FNA":
		return Location{"Freetown", "Africa/Freetown", 8.616440, -13.195500}
	case "FNC":
		return Location{"Funchal", "Atlantic/Madeira", 32.697899, -16.774500}
	case "FNI":
		return Location{"Nimes/Garons", "Europe/Paris", 43.757401, 4.416350}
	case "FNT":
		return Location{"Flint", "America/Detroit", 42.965401, -83.743599}
	case "FOC":
		return Location{"Fuzhou", "Asia/Shanghai", 25.935101, 119.663002}
	case "FOD":
		return Location{"Fort Dodge", "America/Chicago", 42.551498, -94.192596}
	case "FOG":
		return Location{"Foggia", "Europe/Rome", 41.432899, 15.535000}
	case "FON":
		return Location{"La Fortuna/San Carlos", "America/Costa_Rica", 10.478000, -84.634499}
	case "FOR":
		return Location{"Fortaleza", "America/Fortaleza", -3.776280, -38.532600}
	case "FPO":
		return Location{"Freeport", "America/Nassau", 26.558701, -78.695602}
	case "FRA":
		return Location{"Frankfurt-am-Main", "Europe/Berlin", 50.026402, 8.543130}
	case "FRD":
		return Location{"Friday Harbor", "America/Los_Angeles", 48.521999, -123.024002}
	case "FRE":
		return Location{"Fera Island", "Pacific/Guadalcanal", -8.107500, 159.576996}
	case "FRL":
		return Location{"Forli", "Europe/Rome", 44.194801, 12.070100}
	case "FRO":
		return Location{"Floro", "Europe/Oslo", 61.583599, 5.024720}
	case "FRS":
		return Location{"San Benito", "America/Guatemala", 16.913799, -89.866402}
	case "FRU":
		return Location{"Bishkek", "Asia/Bishkek", 43.061272, 74.477508}
	case "FRW":
		return Location{"Francistown", "Africa/Gaborone", -21.159599, 27.474501}
	case "FSC":
		return Location{"Figari Sud-Corse", "Europe/Paris", 41.500599, 9.097780}
	case "FSD":
		return Location{"Sioux Falls", "America/Chicago", 43.582001, -96.741898}
	case "FSM":
		return Location{"Fort Smith", "America/Chicago", 35.336601, -94.367401}
	case "FSZ":
		return Location{"", "Asia/Tokyo", 34.796043, 138.187752}
	case "FTA":
		return Location{"Futuna Island", "Pacific/Efate", -19.516399, 170.231995}
	case "FTE":
		return Location{"El Calafate", "America/Argentina/Rio_Gallegos", -50.280300, -72.053101}
	case "FTU":
		return Location{"Tolanaro", "Indian/Antananarivo", -25.038099, 46.956100}
	case "FUE":
		return Location{"Fuerteventura Island", "Atlantic/Canary", 28.452700, -13.863800}
	case "FUG":
		return Location{"Fuyang", "Asia/Shanghai", 32.882157, 115.734364}
	case "FUJ":
		return Location{"Goto", "Asia/Tokyo", 32.666302, 128.832993}
	case "FUK":
		return Location{"Fukuoka", "Asia/Tokyo", 33.585899, 130.451004}
	case "FUN":
		return Location{"Funafuti", "Pacific/Funafuti", -8.525000, 179.195999}
	case "FUO":
		return Location{"Foshan", "Asia/Shanghai", 23.083300, 113.070000}
	case "FVM":
		return Location{"Fuvahmulah Island", "Indian/Maldives", -0.309722, 73.435000}
	case "FWA":
		return Location{"Fort Wayne", "America/Indiana/Indianapolis", 40.978500, -85.195099}
	case "FYU":
		return Location{"Fort Yukon", "America/Anchorage", 66.571503, -145.250000}
	case "GAE":
		return Location{"Gabes", "Africa/Tunis", 33.876900, 10.103300}
	case "GAJ":
		return Location{"Yamagata", "Asia/Tokyo", 38.411900, 140.371002}
	case "GAL":
		return Location{"Galena", "America/Anchorage", 64.736198, -156.936996}
	case "GAM":
		return Location{"Gambell", "America/Nome", 63.766800, -171.733002}
	case "GAN":
		return Location{"Gan", "Indian/Maldives", -0.693342, 73.155602}
	case "GAU":
		return Location{"Guwahati", "Asia/Kolkata", 26.106100, 91.585899}
	case "GAY":
		return Location{"", "Asia/Kolkata", 24.744301, 84.951202}
	case "GBE":
		return Location{"Gaborone", "Africa/Gaborone", -24.555201, 25.918200}
	case "GBI":
		return Location{"Grand Bahama", "America/Nassau", 26.631901, -78.359200}
	case "GBT":
		return Location{"Gorgan", "Asia/Tehran", 36.909401, 54.401299}
	case "GCC":
		return Location{"Gillette", "America/Denver", 44.348900, -105.539001}
	case "GCI":
		return Location{"Saint Peter Port", "Europe/Guernsey", 49.435001, -2.601970}
	case "GCK":
		return Location{"Garden City", "America/Chicago", 37.927502, -100.723999}
	case "GCM":
		return Location{"Georgetown", "America/Cayman", 19.292801, -81.357697}
	case "GCN":
		return Location{"Grand Canyon", "America/Phoenix", 35.952400, -112.147003}
	case "GCW":
		return Location{"Peach Springs", "America/Phoenix", 35.990398, -113.816002}
	case "GDE":
		return Location{"Gode", "Africa/Addis_Ababa", 5.935130, 43.578602}
	case "GDL":
		return Location{"Guadalajara", "America/Mexico_City", 20.521799, -103.310997}
	case "GDN":
		return Location{"Gdansk", "Europe/Warsaw", 54.377602, 18.466200}
	case "GDQ":
		return Location{"Gondar", "Africa/Addis_Ababa", 12.519900, 37.433998}
	case "GDT":
		return Location{"Cockburn Town", "America/Grand_Turk", 21.444500, -71.142303}
	case "GDV":
		return Location{"Glendive", "America/Denver", 47.138699, -104.806999}
	case "GDX":
		return Location{"Magadan", "Asia/Magadan", 59.910999, 150.720001}
	case "GEA":
		return Location{"Noumea", "Pacific/Noumea", -22.258301, 166.473007}
	case "GEG":
		return Location{"Spokane", "America/Los_Angeles", 47.619900, -117.533997}
	case "GEL":
		return Location{"Santo Angelo", "America/Sao_Paulo", -28.281700, -54.169102}
	case "GEO":
		return Location{"Georgetown", "America/Guyana", 6.498550, -58.254101}
	case "GES":
		return Location{"General Santos City", "Asia/Manila", 6.058000, 125.096001}
	case "GET":
		return Location{"", "Australia/Perth", -28.796101, 114.707001}
	case "GEV":
		return Location{"Gallivare", "Europe/Stockholm", 67.132401, 20.814600}
	case "GFF":
		return Location{"Griffith", "Australia/Sydney", -34.250801, 146.067001}
	case "GFK":
		return Location{"Grand Forks", "America/Chicago", 47.949299, -97.176102}
	case "GGF":
		return Location{"Almeirim", "America/Santarem", -1.491944, -52.578335}
	case "GGG":
		return Location{"Longview", "America/Chicago", 32.383999, -94.711502}
	case "GGM":
		return Location{"Kakamega", "Africa/Nairobi", 0.271342, 34.787300}
	case "GGT":
		return Location{"George Town", "America/Nassau", 23.562599, -75.877998}
	case "GGW":
		return Location{"Glasgow", "America/Denver", 48.212502, -106.614998}
	case "GHA":
		return Location{"Ghardaia", "Africa/Algiers", 32.384102, 3.794110}
	case "GHB":
		return Location{"Governor's Harbour", "America/Nassau", 25.284700, -76.331001}
	case "GHC":
		return Location{"", "America/Nassau", 25.738300, -77.840103}
	case "GHT":
		return Location{"Ghat", "Africa/Tripoli", 25.145599, 10.142600}
	case "GIB":
		return Location{"Gibraltar", "Europe/Gibraltar", 36.151199, -5.349660}
	case "GIC":
		return Location{"", "Australia/Brisbane", -9.232780, 142.218002}
	case "GIG":
		return Location{"Rio De Janeiro", "America/Sao_Paulo", -22.809999, -43.250557}
	case "GIL":
		return Location{"Gilgit", "Asia/Karachi", 35.918800, 74.333603}
	case "GIS":
		return Location{"Gisborne", "Pacific/Auckland", -38.663300, 177.977997}
	case "GIU":
		return Location{"Sigiriya", "Asia/Colombo", 7.956670, 80.728500}
	case "GIZ":
		return Location{"Jizan", "Asia/Riyadh", 16.901100, 42.585800}
	case "GJA":
		return Location{"Guanaja", "America/Tegucigalpa", 16.445400, -85.906601}
	case "GJL":
		return Location{"Jijel", "Africa/Algiers", 36.795101, 5.873610}
	case "GJT":
		return Location{"Grand Junction", "America/Denver", 39.122398, -108.527000}
	case "GKA":
		return Location{"Goronka", "Pacific/Port_Moresby", -6.081690, 145.391998}
	case "GLA":
		return Location{"Glasgow", "Europe/London", 55.871899, -4.433060}
	case "GLF":
		return Location{"Golfito", "America/Costa_Rica", 8.654010, -83.182198}
	case "GLH":
		return Location{"Greenville", "America/Chicago", 33.482899, -90.985603}
	case "GLT":
		return Location{"Gladstone", "Australia/Brisbane", -23.869699, 151.223007}
	case "GLV":
		return Location{"Golovin", "America/Nome", 64.550499, -163.007004}
	case "GMA":
		return Location{"Gemena", "Africa/Kinshasa", 3.235370, 19.771299}
	case "GMB":
		return Location{"Gambela", "Africa/Addis_Ababa", 8.128760, 34.563099}
	case "GMO":
		return Location{"Gombe", "Africa/Lagos", 10.298333, 10.896389}
	case "GMP":
		return Location{"Seoul", "Asia/Seoul", 37.558300, 126.791000}
	case "GMR":
		return Location{"", "Pacific/Gambier", -23.079901, -134.889999}
	case "GMZ":
		return Location{"Alajero", "Atlantic/Canary", 28.029600, -17.214600}
	case "GNB":
		return Location{"Grenoble/Saint-Geoirs", "Europe/Paris", 45.362900, 5.329370}
	case "GND":
		return Location{"Saint George's", "America/Grenada", 12.004200, -61.786201}
	case "GNJ":
		return Location{"Ganja", "Asia/Baku", 40.737701, 46.317600}
	case "GNM":
		return Location{"Guanambi", "America/Bahia", -14.208200, -42.746101}
	case "GNS":
		return Location{"Gunung Sitoli-Nias Island", "Asia/Jakarta", 1.166380, 97.704697}
	case "GNV":
		return Location{"Gainesville", "America/New_York", 29.690100, -82.271797}
	case "GNY":
		return Location{"Sanliurfa", "Europe/Istanbul", 37.445663, 38.895592}
	case "GOA":
		return Location{"Genova", "Europe/Rome", 44.413300, 8.837500}
	case "GOB":
		return Location{"Goba", "Africa/Addis_Ababa", 7.017000, 40.000000}
	case "GOH":
		return Location{"Nuuk", "America/Nuuk", 64.190903, -51.678101}
	case "GOI":
		return Location{"Dabolim", "Asia/Kolkata", 15.380833, 73.827000}
	case "GOJ":
		return Location{"Nizhny Novgorod", "Europe/Moscow", 56.230099, 43.784000}
	case "GOM":
		return Location{"Goma", "Africa/Kigali", -1.670810, 29.238501}
	case "GOP":
		return Location{"Gorakhpur", "Asia/Kolkata", 26.739700, 83.449699}
	case "GOQ":
		return Location{"Golmud", "Asia/Shanghai", 36.400600, 94.786102}
	case "GOT":
		return Location{"Gothenburg", "Europe/Stockholm", 57.662800, 12.279800}
	case "GOU":
		return Location{"Garoua", "Africa/Douala", 9.335890, 13.370100}
	case "GOV":
		return Location{"Nhulunbuy", "Australia/Darwin", -12.269400, 136.817993}
	case "GOY":
		return Location{"Amparai", "Asia/Colombo", 7.337080, 81.625900}
	case "GPA":
		return Location{"Patras", "Europe/Athens", 38.151100, 21.425600}
	case "GPB":
		return Location{"Guarapuava", "America/Sao_Paulo", -25.387501, -51.520199}
	case "GPI":
		return Location{"Guapi", "America/Bogota", 2.570130, -77.898600}
	case "GPS":
		return Location{"Baltra", "Pacific/Galapagos", -0.453758, -90.265900}
	case "GPT":
		return Location{"Gulfport", "America/Chicago", 30.407301, -89.070099}
	case "GRB":
		return Location{"Green Bay", "America/Chicago", 44.485100, -88.129601}
	case "GRI":
		return Location{"Grand Island", "America/Chicago", 40.967499, -98.309601}
	case "GRJ":
		return Location{"George", "Africa/Johannesburg", -34.005600, 22.378901}
	case "GRK":
		return Location{"Fort Hood/Killeen", "America/Chicago", 31.067200, -97.828903}
	case "GRO":
		return Location{"Girona", "Europe/Madrid", 41.901001, 2.760550}
	case "GRQ":
		return Location{"Groningen", "Europe/Amsterdam", 53.119701, 6.579440}
	case "GRR":
		return Location{"Grand Rapids", "America/Detroit", 42.880798, -85.522797}
	case "GRU":
		return Location{"Sao Paulo", "America/Sao_Paulo", -23.435556, -46.473057}
	case "GRV":
		return Location{"Grozny", "Europe/Moscow", 43.388302, 45.698601}
	case "GRW":
		return Location{"Santa Cruz da Graciosa", "Atlantic/Azores", 39.092201, -28.029800}
	case "GRX":
		return Location{"Granada", "Europe/Madrid", 37.188702, -3.777360}
	case "GRZ":
		return Location{"Graz", "Europe/Vienna", 46.991100, 15.439600}
	case "GSM":
		return Location{"", "Asia/Tehran", 26.754601, 55.902401}
	case "GSO":
		return Location{"Greensboro", "America/New_York", 36.097801, -79.937302}
	case "GSP":
		return Location{"Greenville", "America/New_York", 34.895699, -82.218903}
	case "GST":
		return Location{"Gustavus", "America/Juneau", 58.425301, -135.707001}
	case "GSV":
		return Location{"Saratov", "Europe/Saratov", 51.712778, 46.171111}
	case "GTE":
		return Location{"Groote Eylandt", "Australia/Darwin", -13.975000, 136.460007}
	case "GTF":
		return Location{"Great Falls", "America/Denver", 47.481998, -111.371002}
	case "GTO":
		return Location{"Gorontalo-Celebes Island", "Asia/Makassar", 0.637119, 122.849998}
	case "GTP":
		return Location{"Grants Pass", "America/Los_Angeles", 42.510101, -123.388000}
	case "GTR":
		return Location{"Columbus/W Point/Starkville", "America/Chicago", 33.450298, -88.591400}
	case "GTS":
		return Location{"", "Australia/Adelaide", -26.948299, 133.606995}
	case "GUA":
		return Location{"Guatemala City", "America/Guatemala", 14.583300, -90.527496}
	case "GUC":
		return Location{"Gunnison", "America/Denver", 38.533901, -106.932999}
	case "GUM":
		return Location{"Hagatna", "Pacific/Guam", 13.483400, 144.796005}
	case "GUP":
		return Location{"Gallup", "America/Denver", 35.511101, -108.789001}
	case "GUR":
		return Location{"Gurney", "Pacific/Port_Moresby", -10.311500, 150.334000}
	case "GUW":
		return Location{"Atyrau", "Asia/Atyrau", 47.121899, 51.821400}
	case "GVA":
		return Location{"Geneva", "Europe/Paris", 46.238098, 6.108950}
	case "GVR":
		return Location{"Governador Valadares", "America/Sao_Paulo", -18.895201, -41.982201}
	case "GWD":
		return Location{"Gwadar", "Asia/Karachi", 25.233299, 62.329498}
	case "GWL":
		return Location{"Gwalior", "Asia/Kolkata", 26.293301, 78.227798}
	case "GWT":
		return Location{"Westerland", "Europe/Berlin", 54.913200, 8.340470}
	case "GXF":
		return Location{"Sayun", "Asia/Aden", 15.966100, 48.788300}
	case "GYD":
		return Location{"Baku", "Asia/Baku", 40.467499, 50.046700}
	case "GYE":
		return Location{"Guayaquil", "America/Guayaquil", -2.157420, -79.883598}
	case "GYN":
		return Location{"Goiania", "America/Sao_Paulo", -16.632000, -49.220699}
	case "GYS":
		return Location{"Guangyuan", "Asia/Shanghai", 32.391102, 105.702003}
	case "GYU":
		return Location{"Guyuan", "Asia/Shanghai", 36.078889, 106.216944}
	case "GZO":
		return Location{"Gizo", "Pacific/Guadalcanal", -8.097780, 156.863998}
	case "GZP":
		return Location{"Gazipasa", "Europe/Istanbul", 36.299217, 32.300598}
	case "GZT":
		return Location{"Gaziantep", "Europe/Istanbul", 36.947201, 37.478699}
	case "HAA":
		return Location{"Hasvik", "Europe/Oslo", 70.486702, 22.139700}
	case "HAC":
		return Location{"Hachijojima", "Asia/Tokyo", 33.115002, 139.785995}
	case "HAD":
		return Location{"Halmstad", "Europe/Stockholm", 56.691101, 12.820200}
	case "HAH":
		return Location{"Moroni", "Indian/Comoro", -11.533700, 43.271900}
	case "HAJ":
		return Location{"Hannover", "Europe/Berlin", 52.461102, 9.685080}
	case "HAK":
		return Location{"Haikou", "Asia/Shanghai", 19.934900, 110.459000}
	case "HAM":
		return Location{"Hamburg", "Europe/Berlin", 53.630402, 9.988230}
	case "HAN":
		return Location{"Hanoi", "Asia/Bangkok", 21.221201, 105.806999}
	case "HAQ":
		return Location{"Haa Dhaalu Atoll", "Indian/Maldives", 6.744230, 73.170502}
	case "HAS":
		return Location{"", "Asia/Riyadh", 27.437901, 41.686298}
	case "HAU":
		return Location{"Karmoy", "Europe/Oslo", 59.345299, 5.208360}
	case "HAV":
		return Location{"Havana", "America/Havana", 22.989201, -82.409103}
	case "HAY":
		return Location{"Aguachica", "America/Bogota", 8.300000, -73.630600}
	case "HBA":
		return Location{"Hobart", "Australia/Hobart", -42.836102, 147.509995}
	case "HBE":
		return Location{"Alexandria", "Africa/Cairo", 30.917700, 29.696400}
	case "HBX":
		return Location{"Hubli", "Asia/Kolkata", 15.361700, 75.084900}
	case "HCQ":
		return Location{"", "Australia/Perth", -18.233900, 127.669998}
	case "HCR":
		return Location{"Holy Cross", "America/Anchorage", 62.188301, -159.774994}
	case "HDF":
		return Location{"Heringsdorf", "Europe/Berlin", 53.878700, 14.152300}
	case "HDG":
		return Location{"Handan", "Asia/Shanghai", 36.525833, 114.425556}
	case "HDK":
		return Location{"Kulhudhuffushi", "Indian/Maldives", 6.631111, 73.066667}
	case "HDN":
		return Location{"Hayden", "America/Denver", 40.481201, -107.218002}
	case "HDS":
		return Location{"Hoedspruit", "Africa/Johannesburg", -24.368601, 31.048700}
	case "HDY":
		return Location{"Hat Yai", "Asia/Bangkok", 6.933210, 100.392998}
	case "HEA":
		return Location{"", "Asia/Kabul", 34.209999, 62.228298}
	case "HEH":
		return Location{"Heho", "Asia/Yangon", 20.747000, 96.792000}
	case "HEK":
		return Location{"Heihe", "Asia/Shanghai", 50.171621, 127.308884}
	case "HEL":
		return Location{"Helsinki", "Europe/Helsinki", 60.317200, 24.963301}
	case "HER":
		return Location{"Heraklion", "Europe/Athens", 35.339699, 25.180300}
	case "HET":
		return Location{"Hohhot", "Asia/Shanghai", 40.851398, 111.823997}
	case "HFE":
		return Location{"Hefei", "Asia/Shanghai", 31.988980, 116.963800}
	case "HFS":
		return Location{"", "Europe/Stockholm", 60.020100, 13.578900}
	case "HFT":
		return Location{"Hammerfest", "Europe/Oslo", 70.679703, 23.668600}
	case "HGA":
		return Location{"Hargeisa", "Africa/Mogadishu", 9.518170, 44.088799}
	case "HGD":
		return Location{"", "Australia/Brisbane", -20.815001, 144.225006}
	case "HGH":
		return Location{"Hangzhou", "Asia/Shanghai", 30.229500, 120.433998}
	case "HGN":
		return Location{"", "Asia/Bangkok", 19.301300, 97.975800}
	case "HGO":
		return Location{"", "Africa/Abidjan", 9.387180, -5.556660}
	case "HGR":
		return Location{"Hagerstown", "America/New_York", 39.707901, -77.729500}
	case "HGU":
		return Location{"Mount Hagen", "Pacific/Port_Moresby", -5.826790, 144.296005}
	case "HHH":
		return Location{"Hilton Head Island", "America/New_York", 32.224400, -80.697502}
	case "HHN":
		return Location{"Hahn", "Europe/Berlin", 49.948700, 7.263890}
	case "HHQ":
		return Location{"Hua Hin", "Asia/Bangkok", 12.636200, 99.951500}
	case "HHR":
		return Location{"Hawthorne", "America/Los_Angeles", 33.922798, -118.334999}
	case "HHZ":
		return Location{"Hikueru Atoll", "Pacific/Tahiti", -17.544701, -142.613998}
	case "HIA":
		return Location{"Huai'an", "Asia/Shanghai", 33.790833, 119.125000}
	case "HIB":
		return Location{"Hibbing", "America/Chicago", 47.386600, -92.838997}
	case "HID":
		return Location{"Horn Island", "Australia/Brisbane", -10.586400, 142.289993}
	case "HIJ":
		return Location{"Hiroshima", "Asia/Tokyo", 34.436100, 132.919006}
	case "HIN":
		return Location{"Sacheon", "Asia/Seoul", 35.088501, 128.070007}
	case "HIR":
		return Location{"Honiara", "Pacific/Guadalcanal", -9.428000, 160.054993}
	case "HJJ":
		return Location{"Huaihua", "Asia/Shanghai", 27.441111, 109.700000}
	case "HJR":
		return Location{"Khajuraho", "Asia/Kolkata", 24.817200, 79.918602}
	case "HKD":
		return Location{"Hakodate", "Asia/Tokyo", 41.770000, 140.822006}
	case "HKG":
		return Location{"Hong Kong", "Asia/Hong_Kong", 22.308901, 113.915001}
	case "HKK":
		return Location{"", "Pacific/Auckland", -42.713600, 170.985001}
	case "HKN":
		return Location{"Hoskins", "Pacific/Port_Moresby", -5.462170, 150.404999}
	case "HKT":
		return Location{"Phuket", "Asia/Bangkok", 8.113200, 98.316902}
	case "HLA":
		return Location{"Johannesburg", "Africa/Johannesburg", -25.938499, 27.926100}
	case "HLD":
		return Location{"Hailar", "Asia/Shanghai", 49.205002, 119.824997}
	case "HLH":
		return Location{"Ulanhot", "Asia/Shanghai", 46.195333, 122.008333}
	case "HLN":
		return Location{"Helena", "America/Denver", 46.606800, -111.983002}
	case "HLP":
		return Location{"Jakarta", "Asia/Jakarta", -6.266610, 106.890999}
	case "HLZ":
		return Location{"Hamilton", "Pacific/Auckland", -37.866699, 175.332001}
	case "HMA":
		return Location{"Khanty-Mansiysk", "Asia/Yekaterinburg", 61.028500, 69.086098}
	case "HMB":
		return Location{"Sohag", "Africa/Cairo", 26.342778, 31.742778}
	case "HME":
		return Location{"Hassi Messaoud", "Africa/Algiers", 31.673000, 6.140440}
	case "HMI":
		return Location{"Hami", "Asia/Shanghai", 42.841400, 93.669197}
	case "HMO":
		return Location{"Hermosillo", "America/Hermosillo", 29.095900, -111.047997}
	case "HMV":
		return Location{"", "Europe/Stockholm", 65.806099, 15.082800}
	case "HNA":
		return Location{"", "Asia/Tokyo", 39.428600, 141.134995}
	case "HND":
		return Location{"Tokyo", "Asia/Tokyo", 35.552299, 139.779999}
	case "HNH":
		return Location{"Hoonah", "America/Juneau", 58.096100, -135.410004}
	case "HNL":
		return Location{"Honolulu", "Pacific/Honolulu", 21.318701, -157.921997}
	case "HNM":
		return Location{"Hana", "Pacific/Honolulu", 20.795601, -156.014008}
	case "HNS":
		return Location{"Haines", "America/Juneau", 59.243801, -135.524002}
	case "HNY":
		return Location{"Hengyang", "Asia/Shanghai", 26.905300, 112.627998}
	case "HOB":
		return Location{"Hobbs", "America/Denver", 32.687500, -103.217003}
	case "HOF":
		return Location{"", "Asia/Riyadh", 25.285299, 49.485199}
	case "HOG":
		return Location{"Holguin", "America/Havana", 20.785601, -76.315102}
	case "HOI":
		return Location{"", "Pacific/Tahiti", -18.074800, -140.945999}
	case "HOM":
		return Location{"Homer", "America/Anchorage", 59.645599, -151.477005}
	case "HOR":
		return Location{"Horta", "Atlantic/Azores", 38.519901, -28.715900}
	case "HOT":
		return Location{"Hot Springs", "America/Chicago", 34.478001, -93.096199}
	case "HOU":
		return Location{"Houston", "America/Chicago", 29.645399, -95.278900}
	case "HOV":
		return Location{"Orsta", "Europe/Oslo", 62.180000, 6.074100}
	case "HOX":
		return Location{"Hommalinn", "Asia/Yangon", 24.899599, 94.914001}
	case "HPA":
		return Location{"Lifuka", "Pacific/Tongatapu", -19.777000, -174.341003}
	case "HPB":
		return Location{"Hooper Bay", "America/Nome", 61.523899, -166.147003}
	case "HPH":
		return Location{"Haiphong", "Asia/Bangkok", 20.819401, 106.724998}
	case "HPN":
		return Location{"White Plains", "America/New_York", 41.067001, -73.707603}
	case "HRB":
		return Location{"Harbin", "Asia/Shanghai", 45.623402, 126.250000}
	case "HRE":
		return Location{"Harare", "Africa/Harare", -17.931801, 31.092800}
	case "HRG":
		return Location{"Hurghada", "Africa/Cairo", 27.178301, 33.799400}
	case "HRL":
		return Location{"Harlingen", "America/Chicago", 26.228500, -97.654404}
	case "HRO":
		return Location{"Harrison", "America/Chicago", 36.261501, -93.154701}
	case "HSA":
		return Location{"Turkistan", "Asia/Almaty", 43.276901, 68.190399}
	case "HSG":
		return Location{"Saga", "Asia/Tokyo", 33.149700, 130.302002}
	case "HSL":
		return Location{"Huslia", "America/Anchorage", 65.697899, -156.350998}
	case "HSN":
		return Location{"Zhoushan", "Asia/Shanghai", 29.934200, 122.362000}
	case "HSV":
		return Location{"Huntsville", "America/Chicago", 34.637199, -86.775101}
	case "HTA":
		return Location{"Chita", "Asia/Chita", 52.026299, 113.306000}
	case "HTG":
		return Location{"Khatanga", "Asia/Krasnoyarsk", 71.978104, 102.490997}
	case "HTI":
		return Location{"Hamilton Island", "Australia/Lindeman", -20.358101, 148.951996}
	case "HTN":
		return Location{"Hotan", "Asia/Shanghai", 37.038502, 79.864899}
	case "HTS":
		return Location{"Huntington", "America/New_York", 38.366699, -82.557999}
	case "HTY":
		return Location{"Hatay", "Europe/Istanbul", 36.362778, 36.282223}
	case "HUH":
		return Location{"Fare", "Pacific/Tahiti", -16.687201, -151.022003}
	case "HUI":
		return Location{"Hue", "Asia/Ho_Chi_Minh", 16.401501, 107.703003}
	case "HUN":
		return Location{"Hualien City", "Asia/Taipei", 24.023100, 121.617996}
	case "HUS":
		return Location{"Hughes", "America/Anchorage", 66.041100, -154.263001}
	case "HUU":
		return Location{"Huanuco", "America/Lima", -9.878810, -76.204803}
	case "HUX":
		return Location{"Huatulco", "America/Mexico_City", 15.775300, -96.262604}
	case "HUY":
		return Location{"Grimsby", "Europe/London", 53.574402, -0.350833}
	case "HUZ":
		return Location{"Huizhou", "Asia/Shanghai", 23.049999, 114.599998}
	case "HVB":
		return Location{"Hervey Bay", "Australia/Brisbane", -25.318899, 152.880005}
	case "HVD":
		return Location{"Khovd", "Asia/Hovd", 47.954102, 91.628197}
	case "HVG":
		return Location{"Honningsvag", "Europe/Oslo", 71.009697, 25.983601}
	case "HVN":
		return Location{"New Haven", "America/New_York", 41.263699, -72.886803}
	case "HVR":
		return Location{"Havre", "America/Denver", 48.542999, -109.762001}
	case "HWN":
		return Location{"Hwange", "Africa/Harare", -18.629900, 27.021000}
	case "HYA":
		return Location{"Hyannis", "America/New_York", 41.669300, -70.280403}
	case "HYD":
		return Location{"Hyderabad", "Asia/Kolkata", 17.231318, 78.429855}
	case "HYN":
		return Location{"Huangyan", "Asia/Shanghai", 28.562201, 121.429001}
	case "HYS":
		return Location{"Hays", "America/Chicago", 38.842201, -99.273201}
	case "HZG":
		return Location{"Hanzhong", "Asia/Shanghai", 33.063599, 107.008003}
	case "IAA":
		return Location{"Igarka", "Asia/Krasnoyarsk", 67.437202, 86.621902}
	case "IAD":
		return Location{"Dulles", "America/New_York", 38.944500, -77.455803}
	case "IAG":
		return Location{"Niagara Falls", "America/New_York", 43.107300, -78.946198}
	case "IAH":
		return Location{"Houston", "America/Chicago", 29.984400, -95.341400}
	case "IAM":
		return Location{"Amenas", "Africa/Algiers", 28.051500, 9.642910}
	case "IAN":
		return Location{"Kiana", "America/Anchorage", 66.975998, -160.436996}
	case "IAO":
		return Location{"Del Carmen", "Asia/Manila", 9.859100, 126.014000}
	case "IAR":
		return Location{"", "Europe/Moscow", 57.560699, 40.157398}
	case "IAS":
		return Location{"Iasi", "Europe/Bucharest", 47.178501, 27.620600}
	case "IBA":
		return Location{"Ibadan", "Africa/Lagos", 7.362460, 3.978330}
	case "IBE":
		return Location{"Ibague", "America/Bogota", 4.421610, -75.133300}
	case "IBR":
		return Location{"Omitama", "Asia/Tokyo", 36.181099, 140.414993}
	case "IBZ":
		return Location{"Ibiza", "Europe/Madrid", 38.872898, 1.373120}
	case "ICI":
		return Location{"Cicia", "Pacific/Fiji", -17.743299, -179.341995}
	case "ICN":
		return Location{"Seoul", "Asia/Seoul", 37.469101, 126.450996}
	case "ICT":
		return Location{"Wichita", "America/Chicago", 37.649899, -97.433098}
	case "IDA":
		return Location{"Idaho Falls", "America/Boise", 43.514599, -112.070999}
	case "IDR":
		return Location{"Indore", "Asia/Kolkata", 22.721800, 75.801102}
	case "IEG":
		return Location{"Babimost", "Europe/Warsaw", 52.138500, 15.798600}
	case "IEV":
		return Location{"Kiev", "Europe/Kiev", 50.401699, 30.449699}
	case "IFJ":
		return Location{"Isafjordur", "Atlantic/Reykjavik", 66.058098, -23.135300}
	case "IFN":
		return Location{"Isfahan", "Asia/Tehran", 32.750801, 51.861301}
	case "IFU":
		return Location{"Ifuru", "Indian/Maldives", 5.708300, 73.025000}
	case "IGA":
		return Location{"Matthew Town", "America/Nassau", 20.975000, -73.666901}
	case "IGD":
		return Location{"Igdir", "Europe/Istanbul", 39.976627, 43.876648}
	case "IGG":
		return Location{"Igiugig", "America/Anchorage", 59.324001, -155.901993}
	case "IGR":
		return Location{"Puerto Iguazu", "America/Argentina/Cordoba", -25.737301, -54.473400}
	case "IGT":
		return Location{"Magas", "Europe/Moscow", 43.322300, 45.012600}
	case "IGU":
		return Location{"Foz Do Iguacu", "America/Argentina/Cordoba", -25.600279, -54.485001}
	case "IIL":
		return Location{"Ilam", "Asia/Tehran", 33.586601, 46.404800}
	case "IJK":
		return Location{"Izhevsk", "Europe/Samara", 56.828098, 53.457500}
	case "IKA":
		return Location{"Tehran", "Asia/Tehran", 35.416100, 51.152199}
	case "IKI":
		return Location{"Iki", "Asia/Tokyo", 33.749001, 129.785004}
	case "IKO":
		return Location{"Nikolski", "America/Nome", 52.941601, -168.848999}
	case "IKS":
		return Location{"Tiksi", "Asia/Yakutsk", 71.697701, 128.903000}
	case "IKT":
		return Location{"Irkutsk", "Asia/Irkutsk", 52.268002, 104.389000}
	case "IKU":
		return Location{"Tamchy", "Asia/Bishkek", 42.585714, 76.701811}
	case "ILD":
		return Location{"Lleida", "Europe/Madrid", 41.728185, 0.535023}
	case "ILG":
		return Location{"Wilmington", "America/New_York", 39.678699, -75.606499}
	case "ILI":
		return Location{"Iliamna", "America/Anchorage", 59.754398, -154.910996}
	case "ILM":
		return Location{"Wilmington", "America/New_York", 34.270599, -77.902603}
	case "ILO":
		return Location{"Iloilo City", "Asia/Manila", 10.833017, 122.493358}
	case "ILP":
		return Location{"Ile des Pins", "Pacific/Noumea", -22.588900, 167.455994}
	case "ILQ":
		return Location{"Ilo", "America/Lima", -17.695000, -71.344002}
	case "ILR":
		return Location{"Ilorin", "Africa/Lagos", 8.440210, 4.493920}
	case "ILY":
		return Location{"Port Ellen", "Europe/London", 55.681900, -6.256670}
	case "IMF":
		return Location{"Imphal", "Asia/Kolkata", 24.760000, 93.896698}
	case "IMP":
		return Location{"Imperatriz", "America/Fortaleza", -5.531290, -47.459999}
	case "IMT":
		return Location{"Iron Mountain Kingsford", "America/Chicago", 45.818401, -88.114502}
	case "INC":
		return Location{"Yinchuan", "Asia/Shanghai", 38.481899, 106.009003}
	case "IND":
		return Location{"Indianapolis", "America/Indiana/Indianapolis", 39.717300, -86.294403}
	case "INF":
		return Location{"In Guezzam", "Africa/Algiers", 19.566999, 5.750000}
	case "INH":
		return Location{"Inhambabe", "Africa/Maputo", -23.876400, 35.408501}
	case "INI":
		return Location{"Nis", "Europe/Belgrade", 43.337299, 21.853701}
	case "INL":
		return Location{"International Falls", "America/Chicago", 48.566200, -93.403099}
	case "INN":
		return Location{"Innsbruck", "Europe/Vienna", 47.260201, 11.344000}
	case "INU":
		return Location{"Yaren District", "Pacific/Nauru", -0.547458, 166.919006}
	case "INV":
		return Location{"Inverness", "Europe/London", 57.542500, -4.047500}
	case "INZ":
		return Location{"In Salah", "Africa/Algiers", 27.250999, 2.512020}
	case "IOA":
		return Location{"Ioannina", "Europe/Athens", 39.696400, 20.822500}
	case "IOM":
		return Location{"Castletown", "Europe/Isle_of_Man", 54.083302, -4.623890}
	case "IOS":
		return Location{"Ilheus", "America/Bahia", -14.816000, -39.033199}
	case "IPA":
		return Location{"Ipota", "Pacific/Efate", -18.856389, 169.283333}
	case "IPC":
		return Location{"Isla De Pascua", "Pacific/Easter", -27.164801, -109.421997}
	case "IPH":
		return Location{"Ipoh", "Asia/Kuala_Lumpur", 4.567970, 101.092003}
	case "IPI":
		return Location{"Ipiales", "America/Bogota", 0.861925, -77.671800}
	case "IPL":
		return Location{"Imperial", "America/Los_Angeles", 32.834202, -115.579002}
	case "IPN":
		return Location{"Ipatinga", "America/Sao_Paulo", -19.470699, -42.487598}
	case "IQM":
		return Location{"Qiemo", "Asia/Shanghai", 38.149399, 85.532799}
	case "IQN":
		return Location{"Qingyang", "Asia/Shanghai", 35.799702, 107.602997}
	case "IQQ":
		return Location{"Iquique", "America/Santiago", -20.535200, -70.181297}
	case "IQT":
		return Location{"Iquitos", "America/Lima", -3.784740, -73.308800}
	case "IRA":
		return Location{"Kirakira", "Pacific/Guadalcanal", -10.449700, 161.897995}
	case "IRC":
		return Location{"Circle", "America/Anchorage", 65.830498, -144.076004}
	case "IRG":
		return Location{"", "Australia/Brisbane", -12.786900, 143.304993}
	case "IRI":
		return Location{"Nduli", "Africa/Dar_es_Salaam", -7.668630, 35.752102}
	case "IRJ":
		return Location{"La Rioja", "America/Argentina/La_Rioja", -29.381599, -66.795799}
	case "IRK":
		return Location{"Kirksville", "America/Chicago", 40.093498, -92.544899}
	case "IRM":
		return Location{"", "Asia/Yekaterinburg", 63.198799, 64.439301}
	case "IRP":
		return Location{"", "Africa/Lubumbashi", 2.827610, 27.588301}
	case "IRZ":
		return Location{"Santa Isabel Do Rio Negro", "America/Manaus", -0.416944, -65.033890}
	case "ISA":
		return Location{"Mount Isa", "Australia/Brisbane", -20.663900, 139.488998}
	case "ISB":
		return Location{"Islamabad", "Asia/Karachi", 33.549083, 72.825650}
	case "ISE":
		return Location{"Isparta", "Europe/Istanbul", 37.855400, 30.368401}
	case "ISG":
		return Location{"Ishigaki", "Asia/Tokyo", 24.344500, 124.186996}
	case "ISK":
		return Location{"Nashik", "Asia/Kolkata", 20.119101, 73.912903}
	case "ISP":
		return Location{"Islip", "America/New_York", 40.795200, -73.100197}
	case "IST":
		return Location{"Arnavutkoy", "Europe/Istanbul", 41.262222, 28.727778}
	case "ISU":
		return Location{"Sulaymaniyah", "Asia/Baghdad", 35.561749, 45.316738}
	case "ITB":
		return Location{"Itaituba", "America/Santarem", -4.242340, -56.000702}
	case "ITH":
		return Location{"Ithaca", "America/New_York", 42.491001, -76.458397}
	case "ITM":
		return Location{"Osaka", "Asia/Tokyo", 34.785500, 135.438004}
	case "ITO":
		return Location{"Hilo", "Pacific/Honolulu", 19.721399, -155.048004}
	case "IUE":
		return Location{"Alofi", "Pacific/Niue", -19.079031, -169.925598}
	case "IVC":
		return Location{"Invercargill", "Pacific/Auckland", -46.412399, 168.313004}
	case "IVL":
		return Location{"Ivalo", "Europe/Helsinki", 68.607300, 27.405300}
	case "IVR":
		return Location{"", "Australia/Sydney", -29.888300, 151.143997}
	case "IWA":
		return Location{"Ivanovo", "Europe/Moscow", 56.939400, 40.940800}
	case "IWD":
		return Location{"Ironwood", "America/Menominee", 46.527500, -90.131401}
	case "IWJ":
		return Location{"Masuda", "Asia/Tokyo", 34.676399, 131.789993}
	case "IWK":
		return Location{"Iwakuni", "Asia/Tokyo", 34.143902, 132.235992}
	case "IXA":
		return Location{"Agartala", "Asia/Dhaka", 23.886999, 91.240402}
	case "IXB":
		return Location{"Siliguri", "Asia/Kolkata", 26.681200, 88.328598}
	case "IXC":
		return Location{"Chandigarh", "Asia/Kolkata", 30.673500, 76.788498}
	case "IXD":
		return Location{"Allahabad", "Asia/Kolkata", 25.440100, 81.733902}
	case "IXE":
		return Location{"Mangalore", "Asia/Kolkata", 12.961300, 74.890099}
	case "IXG":
		return Location{"", "Asia/Kolkata", 15.859300, 74.618301}
	case "IXI":
		return Location{"Lilabari", "Asia/Kolkata", 27.295500, 94.097603}
	case "IXJ":
		return Location{"Jammu", "Asia/Kolkata", 32.689098, 74.837402}
	case "IXK":
		return Location{"", "Asia/Kolkata", 21.317101, 70.270401}
	case "IXL":
		return Location{"Leh", "Asia/Kolkata", 34.135899, 77.546501}
	case "IXM":
		return Location{"Madurai", "Asia/Kolkata", 9.834510, 78.093399}
	case "IXR":
		return Location{"Ranchi", "Asia/Kolkata", 23.314301, 85.321701}
	case "IXS":
		return Location{"Silchar", "Asia/Kolkata", 24.912901, 92.978699}
	case "IXT":
		return Location{"Pasighat", "Asia/Kolkata", 28.066099, 95.335602}
	case "IXU":
		return Location{"Aurangabad", "Asia/Kolkata", 19.862700, 75.398102}
	case "IXW":
		return Location{"", "Asia/Kolkata", 22.813200, 86.168800}
	case "IXY":
		return Location{"Kandla", "Asia/Kolkata", 23.112700, 70.100304}
	case "IXZ":
		return Location{"Port Blair", "Asia/Kolkata", 11.641200, 92.729698}
	case "IZA":
		return Location{"Juiz De Fora", "America/Sao_Paulo", -21.513056, -43.173058}
	case "IZO":
		return Location{"Izumo", "Asia/Tokyo", 35.413601, 132.889999}
	case "JAC":
		return Location{"Jackson", "America/Denver", 43.607300, -110.737999}
	case "JAE":
		return Location{"Jaen", "America/Lima", -5.592480, -78.774002}
	case "JAF":
		return Location{"Jaffna", "Asia/Colombo", 9.792330, 80.070099}
	case "JAI":
		return Location{"Jaipur", "Asia/Kolkata", 26.824200, 75.812202}
	case "JAN":
		return Location{"Jackson", "America/Chicago", 32.311199, -90.075897}
	case "JAU":
		return Location{"Jauja", "America/Lima", -11.783100, -75.473396}
	case "JAV":
		return Location{"Ilulissat", "America/Nuuk", 69.243202, -51.057098}
	case "JAX":
		return Location{"Jacksonville", "America/New_York", 30.494101, -81.687897}
	case "JBQ":
		return Location{"La Isabela", "America/Santo_Domingo", 18.572500, -69.985603}
	case "JBR":
		return Location{"Jonesboro", "America/Chicago", 35.831699, -90.646400}
	case "JCK":
		return Location{"", "Australia/Brisbane", -20.668301, 141.723007}
	case "JDH":
		return Location{"Jodhpur", "Asia/Kolkata", 26.251101, 73.048897}
	case "JDO":
		return Location{"Juazeiro Do Norte", "America/Fortaleza", -7.218960, -39.270100}
	case "JDZ":
		return Location{"Jingdezhen", "Asia/Shanghai", 29.338600, 117.176003}
	case "JED":
		return Location{"Jeddah", "Asia/Riyadh", 21.679600, 39.156502}
	case "JEE":
		return Location{"Jeremie", "America/Port-au-Prince", 18.663099, -74.170303}
	case "JEG":
		return Location{"Aasiaat", "America/Nuuk", 68.721802, -52.784698}
	case "JEK":
		return Location{"Lower Zambezi National Park", "Africa/Harare", -15.633332, 29.603333}
	case "JER":
		return Location{"Saint Helier", "Europe/Jersey", 49.207901, -2.195510}
	case "JFK":
		return Location{"New York", "America/New_York", 40.639801, -73.778900}
	case "JFR":
		return Location{"Paamiut", "America/Nuuk", 62.014736, -49.670937}
	case "JGA":
		return Location{"Jamnagar", "Asia/Kolkata", 22.465500, 70.012604}
	case "JGN":
		return Location{"Jiayuguan", "Asia/Shanghai", 39.856899, 98.341400}
	case "JGS":
		return Location{"Ji'an", "Asia/Shanghai", 26.856899, 114.737000}
	case "JHB":
		return Location{"Senai", "Asia/Kuala_Lumpur", 1.641310, 103.669998}
	case "JHG":
		return Location{"Jinghong", "Asia/Shanghai", 21.973900, 100.760002}
	case "JHM":
		return Location{"Lahaina", "Pacific/Honolulu", 20.962900, -156.673004}
	case "JHS":
		return Location{"Sisimiut", "America/Nuuk", 66.951302, -53.729301}
	case "JIA":
		return Location{"Juina", "America/Cuiaba", -11.419444, -58.701668}
	case "JIB":
		return Location{"Djibouti City", "Africa/Djibouti", 11.547300, 43.159500}
	case "JIC":
		return Location{"Jinchang", "Asia/Shanghai", 38.542222, 102.348333}
	case "JIJ":
		return Location{"Jijiga", "Africa/Addis_Ababa", 9.330833, 42.911111}
	case "JIK":
		return Location{"Ikaria Island", "Europe/Athens", 37.682701, 26.347099}
	case "JIM":
		return Location{"Jimma", "Africa/Addis_Ababa", 7.666090, 36.816601}
	case "JIN":
		return Location{"Jinja", "Africa/Kampala", 0.450000, 33.200001}
	case "JIQ":
		return Location{"Chongqing", "Asia/Shanghai", 29.514559, 108.833715}
	case "JIU":
		return Location{"Jiujiang", "Asia/Shanghai", 29.476944, 115.801111}
	case "JJD":
		return Location{"Jijoca de Jericoacoara", "America/Fortaleza", -2.906667, -40.358056}
	case "JJN":
		return Location{"Quanzhou", "Asia/Shanghai", 24.796400, 118.589996}
	case "JKH":
		return Location{"Chios Island", "Europe/Athens", 38.343201, 26.140600}
	case "JKL":
		return Location{"Kalymnos Island", "Europe/Athens", 36.963299, 26.940599}
	case "JKR":
		return Location{"Janakpur", "Asia/Kathmandu", 26.708799, 85.922401}
	case "JLN":
		return Location{"Joplin", "America/Chicago", 37.151798, -94.498299}
	case "JLR":
		return Location{"", "Asia/Kolkata", 23.177799, 80.052002}
	case "JMK":
		return Location{"Mykonos Island", "Europe/Athens", 37.435101, 25.348101}
	case "JMS":
		return Location{"Jamestown", "America/Chicago", 46.929699, -98.678200}
	case "JMU":
		return Location{"Jiamusi", "Asia/Shanghai", 46.843399, 130.464996}
	case "JNB":
		return Location{"Johannesburg", "Africa/Johannesburg", -26.133333, 28.250000}
	case "JNG":
		return Location{"Jining", "Asia/Shanghai", 35.292778, 116.346667}
	case "JNU":
		return Location{"Juneau", "America/Juneau", 58.355000, -134.576004}
	case "JNX":
		return Location{"Naxos Island", "Europe/Athens", 37.081100, 25.368099}
	case "JNZ":
		return Location{"Jinzhou", "Asia/Shanghai", 41.101398, 121.061996}
	case "JOE":
		return Location{"Joensuu / Liperi", "Europe/Helsinki", 62.662899, 29.607500}
	case "JOG":
		return Location{"Yogyakarta-Java Island", "Asia/Jakarta", -7.788180, 110.431999}
	case "JOI":
		return Location{"Joinville", "America/Sao_Paulo", -26.224501, -48.797401}
	case "JOS":
		return Location{"Jos", "Africa/Lagos", 9.639830, 8.869050}
	case "JPA":
		return Location{"Joao Pessoa", "America/Fortaleza", -7.145833, -34.948612}
	case "JPR":
		return Location{"Ji-Parana", "America/Porto_Velho", -10.870800, -61.846500}
	case "JQA":
		return Location{"Uummannaq", "America/Nuuk", 70.734200, -52.696201}
	case "JRG":
		return Location{"Jharsuguda", "Asia/Kolkata", 21.913500, 84.050400}
	case "JRH":
		return Location{"Jorhat", "Asia/Kolkata", 26.731501, 94.175499}
	case "JRO":
		return Location{"Arusha", "Africa/Dar_es_Salaam", -3.429410, 37.074501}
	case "JSA":
		return Location{"", "Asia/Kolkata", 26.888700, 70.864998}
	case "JSH":
		return Location{"Crete Island", "Europe/Athens", 35.216099, 26.101299}
	case "JSI":
		return Location{"Skiathos", "Europe/Athens", 39.177101, 23.503700}
	case "JSR":
		return Location{"Jashahor", "Asia/Dhaka", 23.183800, 89.160797}
	case "JST":
		return Location{"Johnstown", "America/New_York", 40.316101, -78.833900}
	case "JSU":
		return Location{"Maniitsoq", "America/Nuuk", 65.412498, -52.939400}
	case "JSY":
		return Location{"Syros Island", "Europe/Athens", 37.422798, 24.950899}
	case "JTC":
		return Location{"Bauru", "America/Sao_Paulo", -22.157780, -49.068330}
	case "JTR":
		return Location{"Santorini Island", "Europe/Athens", 36.399200, 25.479300}
	case "JTY":
		return Location{"Astypalaia Island", "Europe/Athens", 36.579899, 26.375799}
	case "JUB":
		return Location{"Juba", "Africa/Juba", 4.872010, 31.601101}
	case "JUJ":
		return Location{"San Salvador de Jujuy", "America/Argentina/Jujuy", -24.392799, -65.097801}
	case "JUL":
		return Location{"Juliaca", "America/Lima", -15.467100, -70.158203}
	case "JUV":
		return Location{"Upernavik", "America/Nuuk", 72.790199, -56.130600}
	case "JUZ":
		return Location{"Quzhou", "Asia/Shanghai", 28.965799, 118.899002}
	case "JYV":
		return Location{"Jyvaskylan Maalaiskunta", "Europe/Helsinki", 62.399502, 25.678301}
	case "JZH":
		return Location{"Jiuzhaigou", "Asia/Shanghai", 32.853333, 103.682222}
	case "KAA":
		return Location{"Kasama", "Africa/Lusaka", -10.216700, 31.133301}
	case "KAB":
		return Location{"Kariba", "Africa/Harare", -16.519800, 28.885000}
	case "KAD":
		return Location{"Kaduna", "Africa/Lagos", 10.696000, 7.320110}
	case "KAJ":
		return Location{"Kajaani", "Europe/Helsinki", 64.285500, 27.692400}
	case "KAL":
		return Location{"Kaltag", "America/Anchorage", 64.319099, -158.740997}
	case "KAN":
		return Location{"Kano", "Africa/Lagos", 12.047600, 8.524620}
	case "KAO":
		return Location{"Kuusamo", "Europe/Helsinki", 65.987602, 29.239401}
	case "KAW":
		return Location{"Kawthoung", "Asia/Yangon", 10.049300, 98.538002}
	case "KAZ":
		return Location{"Kao-Celebes Island", "Asia/Jayapura", 1.185280, 127.896004}
	case "KBH":
		return Location{"Kalat", "Asia/Karachi", 29.133333, 66.516670}
	case "KBL":
		return Location{"Kabul", "Asia/Kabul", 34.565899, 69.212303}
	case "KBP":
		return Location{"Kiev", "Europe/Kiev", 50.345001, 30.894699}
	case "KBR":
		return Location{"Kota Baharu", "Asia/Kuala_Lumpur", 6.166850, 102.292999}
	case "KBU":
		return Location{"Laut Island", "Asia/Makassar", -3.294720, 116.165001}
	case "KBV":
		return Location{"Krabi", "Asia/Bangkok", 8.099120, 98.986198}
	case "KCA":
		return Location{"Kuqa", "Asia/Shanghai", 41.718102, 82.986900}
	case "KCH":
		return Location{"Kuching", "Asia/Kuching", 1.484700, 110.347000}
	case "KCK":
		return Location{"Kirensk", "Asia/Irkutsk", 57.772999, 108.064003}
	case "KCM":
		return Location{"Kahramanmaras", "Europe/Istanbul", 37.538826, 36.953522}
	case "KCT":
		return Location{"Galle", "Asia/Colombo", 5.993680, 80.320297}
	case "KCZ":
		return Location{"Nankoku", "Asia/Tokyo", 33.546101, 133.669006}
	case "KDH":
		return Location{"", "Asia/Kabul", 31.505800, 65.847801}
	case "KDI":
		return Location{"Kendari-Celebes Island", "Asia/Makassar", -4.081610, 122.417999}
	case "KDM":
		return Location{"Huvadhu Atoll", "Indian/Maldives", 0.488131, 72.996902}
	case "KDO":
		return Location{"Kadhdhoo", "Indian/Maldives", 1.859170, 73.521896}
	case "KDU":
		return Location{"Skardu", "Asia/Karachi", 35.335499, 75.536003}
	case "KDV":
		return Location{"Vunisea", "Pacific/Fiji", -19.058100, 178.156998}
	case "KEF":
		return Location{"Reykjavik", "Atlantic/Reykjavik", 63.985001, -22.605600}
	case "KEJ":
		return Location{"Kemerovo", "Asia/Novokuznetsk", 55.270100, 86.107201}
	case "KEM":
		return Location{"Kemi / Tornio", "Europe/Helsinki", 65.778702, 24.582100}
	case "KEO":
		return Location{"Odienne", "Africa/Abidjan", 9.500000, -7.567000}
	case "KEP":
		return Location{"Nepalgunj", "Asia/Kathmandu", 28.103600, 81.667000}
	case "KER":
		return Location{"Kerman", "Asia/Tehran", 30.274401, 56.951099}
	case "KET":
		return Location{"Kengtung", "Asia/Yangon", 21.301600, 99.636002}
	case "KEU":
		return Location{"Keekorok", "Africa/Nairobi", -1.583000, 35.250000}
	case "KFP":
		return Location{"False Pass", "America/Nome", 54.847401, -163.410004}
	case "KFS":
		return Location{"Kastamonu", "Europe/Istanbul", 41.314201, 33.795799}
	case "KGA":
		return Location{"Kananga", "Africa/Lubumbashi", -5.900050, 22.469200}
	case "KGC":
		return Location{"", "Australia/Adelaide", -35.713902, 137.520996}
	case "KGD":
		return Location{"Kaliningrad", "Europe/Kaliningrad", 54.889999, 20.592600}
	case "KGE":
		return Location{"Kagau Island", "Pacific/Guadalcanal", -7.333000, 157.582993}
	case "KGF":
		return Location{"Karaganda", "Asia/Almaty", 49.670799, 73.334396}
	case "KGI":
		return Location{"Kalgoorlie", "Australia/Perth", -30.789400, 121.461998}
	case "KGK":
		return Location{"Koliganek", "America/Anchorage", 59.726601, -157.259003}
	case "KGL":
		return Location{"Kigali", "Africa/Kigali", -1.968630, 30.139500}
	case "KGP":
		return Location{"Kogalym", "Asia/Yekaterinburg", 62.190399, 74.533798}
	case "KGS":
		return Location{"Kos Island", "Europe/Athens", 36.793301, 27.091700}
	case "KHG":
		return Location{"Kashgar", "Asia/Shanghai", 39.542900, 76.019997}
	case "KHH":
		return Location{"Kaohsiung City", "Asia/Taipei", 22.577101, 120.349998}
	case "KHI":
		return Location{"Karachi", "Asia/Karachi", 24.906500, 67.160797}
	case "KHM":
		return Location{"Kanti", "Asia/Yangon", 25.988300, 95.674400}
	case "KHN":
		return Location{"Nanchang", "Asia/Shanghai", 28.865000, 115.900002}
	case "KHS":
		return Location{"Khasab", "Asia/Muscat", 26.171000, 56.240601}
	case "KHT":
		return Location{"Khost", "Asia/Kabul", 33.333401, 69.952003}
	case "KHV":
		return Location{"Khabarovsk", "Asia/Vladivostok", 48.528000, 135.188004}
	case "KID":
		return Location{"Kristianstad", "Europe/Stockholm", 55.921700, 14.085500}
	case "KIE":
		return Location{"Kieta", "Pacific/Bougainville", -6.305000, 155.727778}
	case "KIH":
		return Location{"Kish Island", "Asia/Tehran", 26.526199, 53.980202}
	case "KIJ":
		return Location{"Niigata", "Asia/Tokyo", 37.955898, 139.121002}
	case "KIK":
		return Location{"Kirkuk", "Asia/Baghdad", 35.469501, 44.348900}
	case "KIM":
		return Location{"Kimberley", "Africa/Johannesburg", -28.802799, 24.765200}
	case "KIN":
		return Location{"Kingston", "America/Jamaica", 17.935699, -76.787498}
	case "KIR":
		return Location{"Killarney", "Europe/Dublin", 52.180901, -9.523780}
	case "KIS":
		return Location{"Kisumu", "Africa/Nairobi", -0.086139, 34.728901}
	case "KIT":
		return Location{"Kithira Island", "Europe/Athens", 36.274300, 23.017000}
	case "KIX":
		return Location{"Osaka", "Asia/Tokyo", 34.427299, 135.244003}
	case "KJA":
		return Location{"Krasnoyarsk", "Asia/Krasnoyarsk", 56.172901, 92.493301}
	case "KKA":
		return Location{"Koyuk", "America/Anchorage", 64.939499, -161.154007}
	case "KKC":
		return Location{"Khon Kaen", "Asia/Bangkok", 16.466600, 102.783997}
	case "KKE":
		return Location{"Kerikeri", "Pacific/Auckland", -35.262798, 173.912003}
	case "KKH":
		return Location{"Kongiganak", "America/Nome", 59.960800, -162.880997}
	case "KKJ":
		return Location{"Kitakyushu", "Asia/Tokyo", 33.845901, 131.035004}
	case "KKN":
		return Location{"Kirkenes", "Europe/Oslo", 69.725800, 29.891300}
	case "KKQ":
		return Location{"Krasnoselkup", "Asia/Yekaterinburg", 65.717000, 82.455000}
	case "KKR":
		return Location{"", "Pacific/Tahiti", -15.663300, -146.884995}
	case "KKS":
		return Location{"", "Asia/Tehran", 33.895302, 51.577000}
	case "KKX":
		return Location{"", "Asia/Tokyo", 28.321301, 129.927994}
	case "KLF":
		return Location{"Kaluga", "Europe/Moscow", 54.549999, 36.366669}
	case "KLG":
		return Location{"Kalskag", "America/Anchorage", 61.536301, -160.341003}
	case "KLH":
		return Location{"", "Asia/Kolkata", 16.664700, 74.289398}
	case "KLN":
		return Location{"Larsen Bay", "America/Anchorage", 57.535099, -153.977997}
	case "KLO":
		return Location{"Kalibo", "Asia/Manila", 11.679400, 122.375999}
	case "KLR":
		return Location{"", "Europe/Stockholm", 56.685501, 16.287600}
	case "KLU":
		return Location{"Klagenfurt am Worthersee", "Europe/Vienna", 46.642502, 14.337700}
	case "KLX":
		return Location{"Kalamata", "Europe/Athens", 37.068298, 22.025499}
	case "KME":
		return Location{"Kamembe", "Africa/Kigali", -2.462240, 28.907900}
	case "KMG":
		return Location{"Kunming", "Asia/Shanghai", 24.992399, 102.744003}
	case "KMI":
		return Location{"Miyazaki", "Asia/Tokyo", 31.877199, 131.449005}
	case "KMJ":
		return Location{"Kumamoto", "Asia/Tokyo", 32.837299, 130.854996}
	case "KMO":
		return Location{"Manokotak", "America/Anchorage", 58.990200, -159.050003}
	case "KMQ":
		return Location{"Kanazawa", "Asia/Tokyo", 36.394600, 136.406998}
	case "KMS":
		return Location{"Kumasi", "Africa/Accra", 6.714560, -1.590820}
	case "KMV":
		return Location{"Kalemyo", "Asia/Yangon", 23.188801, 94.051102}
	case "KND":
		return Location{"Kindu", "Africa/Lubumbashi", -2.919180, 25.915400}
	case "KNG":
		return Location{"Kaimana-Papua Island", "Asia/Jayapura", -3.644520, 133.695999}
	case "KNH":
		return Location{"Shang-I", "Asia/Taipei", 24.427900, 118.359001}
	case "KNK":
		return Location{"Kokhanok", "America/Anchorage", 59.433201, -154.804001}
	case "KNO":
		return Location{"Medan-Sumatra Island", "Asia/Jakarta", 3.558060, 98.671700}
	case "KNQ":
		return Location{"Kone", "Pacific/Noumea", -21.054300, 164.837006}
	case "KNS":
		return Location{"", "Australia/Currie", -39.877499, 143.878006}
	case "KNU":
		return Location{"", "Asia/Kolkata", 26.441401, 80.364899}
	case "KNW":
		return Location{"New Stuyahok", "America/Anchorage", 59.449902, -157.328003}
	case "KNX":
		return Location{"Kununurra", "Australia/Perth", -15.778100, 128.707993}
	case "KOA":
		return Location{"Kailua/Kona", "Pacific/Honolulu", 19.738800, -156.046005}
	case "KOE":
		return Location{"Kupang-Timor Island", "Asia/Makassar", -10.171600, 123.670998}
	case "KOI":
		return Location{"Orkney Islands", "Europe/London", 58.957802, -2.905000}
	case "KOJ":
		return Location{"Kagoshima", "Asia/Tokyo", 31.803400, 130.718994}
	case "KOK":
		return Location{"Kokkola / Kruunupyy", "Europe/Helsinki", 63.721199, 23.143101}
	case "KOO":
		return Location{"Kongolo", "Africa/Lubumbashi", -5.394440, 26.990000}
	case "KOP":
		return Location{"", "Asia/Bangkok", 17.383801, 104.642998}
	case "KOS":
		return Location{"Sihanukville", "Asia/Phnom_Penh", 10.579700, 103.637001}
	case "KOT":
		return Location{"Kotlik", "America/Nome", 63.030602, -163.533005}
	case "KOV":
		return Location{"Kokshetau", "Asia/Almaty", 53.329102, 69.594597}
	case "KOW":
		return Location{"Ganzhou", "Asia/Shanghai", 25.825800, 114.912003}
	case "KPN":
		return Location{"Kipnuk", "America/Nome", 59.932999, -164.031006}
	case "KPO":
		return Location{"Pohang", "Asia/Seoul", 35.987900, 129.419998}
	case "KPV":
		return Location{"Perryville", "America/Anchorage", 55.905998, -159.162994}
	case "KPW":
		return Location{"Keperveem", "Asia/Anadyr", 67.845001, 166.139999}
	case "KQH":
		return Location{"Kishangarh", "Asia/Kolkata", 26.591167, 74.816167}
	case "KQT":
		return Location{"Kurgan-Tyube", "Asia/Dushanbe", 37.866402, 68.864700}
	case "KRF":
		return Location{"Kramfors / Solleftea", "Europe/Stockholm", 63.048599, 17.768900}
	case "KRK":
		return Location{"Krakow", "Europe/Warsaw", 50.077702, 19.784800}
	case "KRL":
		return Location{"Korla", "Asia/Shanghai", 41.697800, 86.128899}
	case "KRN":
		return Location{"", "Europe/Stockholm", 67.821999, 20.336800}
	case "KRO":
		return Location{"Kurgan", "Asia/Yekaterinburg", 55.475300, 65.415604}
	case "KRS":
		return Location{"Kjevik", "Europe/Oslo", 58.204201, 8.085370}
	case "KRT":
		return Location{"Khartoum", "Africa/Khartoum", 15.589500, 32.553200}
	case "KRW":
		return Location{"Krasnovodsk", "Asia/Ashgabat", 40.063301, 53.007198}
	case "KRY":
		return Location{"Karamay", "Asia/Shanghai", 45.466550, 84.952700}
	case "KSA":
		return Location{"Okat", "Pacific/Kosrae", 5.356980, 162.957993}
	case "KSC":
		return Location{"Kosice", "Europe/Bratislava", 48.663101, 21.241100}
	case "KSE":
		return Location{"Kasese", "Africa/Kampala", 0.183000, 30.100000}
	case "KSF":
		return Location{"Kassel", "Europe/Berlin", 51.408333, 9.377500}
	case "KSH":
		return Location{"Kermanshah", "Asia/Tehran", 34.345901, 47.158100}
	case "KSJ":
		return Location{"Kasos Island", "Europe/Athens", 35.421398, 26.910000}
	case "KSM":
		return Location{"St Mary's", "America/Nome", 62.060501, -163.302002}
	case "KSN":
		return Location{"Kostanay", "Asia/Qostanay", 53.206902, 63.550301}
	case "KSO":
		return Location{"Kastoria", "Europe/Athens", 40.446301, 21.282200}
	case "KSQ":
		return Location{"Karshi", "Asia/Samarkand", 38.802500, 65.773056}
	case "KSU":
		return Location{"Kvernberget", "Europe/Oslo", 63.111801, 7.824520}
	case "KSY":
		return Location{"Kars", "Europe/Istanbul", 40.562199, 43.115002}
	case "KSZ":
		return Location{"Kotlas", "Europe/Moscow", 61.235802, 46.697498}
	case "KTA":
		return Location{"Karratha", "Australia/Perth", -20.712200, 116.773003}
	case "KTG":
		return Location{"Ketapang-Borneo Island", "Asia/Pontianak", -1.816640, 109.962997}
	case "KTL":
		return Location{"Kitale", "Africa/Nairobi", 0.971989, 34.958599}
	case "KTM":
		return Location{"Kathmandu", "Asia/Kathmandu", 27.696600, 85.359100}
	case "KTN":
		return Location{"Ketchikan", "America/Sitka", 55.355598, -131.714004}
	case "KTR":
		return Location{"", "Australia/Darwin", -14.521100, 132.378006}
	case "KTS":
		return Location{"Brevig Mission", "America/Nome", 65.331299, -166.466003}
	case "KTT":
		return Location{"Kittila", "Europe/Helsinki", 67.700996, 24.846800}
	case "KTW":
		return Location{"Katowice", "Europe/Warsaw", 50.474300, 19.080000}
	case "KUA":
		return Location{"Kuantan", "Asia/Kuala_Lumpur", 3.775390, 103.209000}
	case "KUF":
		return Location{"Samara", "Europe/Samara", 53.504902, 50.164299}
	case "KUG":
		return Location{"", "Australia/Brisbane", -10.225000, 142.218002}
	case "KUH":
		return Location{"Kushiro", "Asia/Tokyo", 43.041000, 144.192993}
	case "KUK":
		return Location{"Kasigluk", "America/Nome", 60.874401, -162.524002}
	case "KUL":
		return Location{"Kuala Lumpur", "Asia/Kuala_Lumpur", 2.745580, 101.709999}
	case "KUM":
		return Location{"", "Asia/Tokyo", 30.385599, 130.658997}
	case "KUN":
		return Location{"Kaunas", "Europe/Vilnius", 54.963902, 24.084801}
	case "KUO":
		return Location{"Kuopio / Siilinjarvi", "Europe/Helsinki", 63.007099, 27.797800}
	case "KUS":
		return Location{"Kulusuk", "America/Nuuk", 65.573601, -37.123600}
	case "KUT":
		return Location{"Kutaisi", "Asia/Tbilisi", 42.176701, 42.482601}
	case "KUU":
		return Location{"", "Asia/Kolkata", 31.876699, 77.154404}
	case "KUV":
		return Location{"Kunsan", "Asia/Seoul", 35.903801, 126.615997}
	case "KVA":
		return Location{"Kavala", "Europe/Athens", 40.913300, 24.619200}
	case "KVC":
		return Location{"King Cove", "America/Nome", 55.116299, -162.266006}
	case "KVG":
		return Location{"Kavieng", "Pacific/Port_Moresby", -2.579400, 150.807999}
	case "KVK":
		return Location{"Apatity", "Europe/Moscow", 67.463303, 33.588299}
	case "KVL":
		return Location{"Kivalina", "America/Nome", 67.736198, -164.563004}
	case "KVX":
		return Location{"Kirov", "Europe/Kirov", 58.503300, 49.348301}
	case "KWA":
		return Location{"Kwajalein", "Pacific/Kwajalein", 8.720120, 167.731995}
	case "KWE":
		return Location{"Guiyang", "Asia/Shanghai", 26.538500, 106.801003}
	case "KWI":
		return Location{"Kuwait City", "Asia/Kuwait", 29.226601, 47.968899}
	case "KWJ":
		return Location{"Gwangju", "Asia/Seoul", 35.126400, 126.808998}
	case "KWK":
		return Location{"Kwigillingok", "America/Nome", 59.876499, -163.169006}
	case "KWL":
		return Location{"Guilin City", "Asia/Shanghai", 25.218100, 110.039001}
	case "KWM":
		return Location{"Kowanyama", "Australia/Brisbane", -15.485600, 141.751007}
	case "KWN":
		return Location{"Quinhagak", "America/Anchorage", 59.755100, -161.845001}
	case "KWT":
		return Location{"Kwethluk", "America/Anchorage", 60.790298, -161.444000}
	case "KWZ":
		return Location{"", "Africa/Lubumbashi", -10.765900, 25.505699}
	case "KXB":
		return Location{"Kolaka", "Asia/Makassar", -4.341217, 121.523983}
	case "KXF":
		return Location{"Koro Island", "Pacific/Fiji", -17.345800, 179.421997}
	case "KXU":
		return Location{"Katiu", "Pacific/Tahiti", -16.339399, -144.403000}
	case "KYA":
		return Location{"Konya", "Europe/Istanbul", 37.979000, 32.561901}
	case "KYK":
		return Location{"Karluk", "America/Anchorage", 57.567101, -154.449997}
	case "KYP":
		return Location{"Kyaukpyu", "Asia/Yangon", 19.426399, 93.534798}
	case "KYU":
		return Location{"Koyukuk", "America/Anchorage", 64.876099, -157.727005}
	case "KYZ":
		return Location{"Kyzyl", "Asia/Krasnoyarsk", 51.669399, 94.400597}
	case "KZI":
		return Location{"Kozani", "Europe/Athens", 40.286098, 21.840799}
	case "KZN":
		return Location{"Kazan", "Europe/Moscow", 55.606201, 49.278702}
	case "KZO":
		return Location{"Kzyl-Orda", "Asia/Qyzylorda", 44.706902, 65.592499}
	case "KZR":
		return Location{"Altintas", "Europe/Istanbul", 39.109025, 30.137130}
	case "KZS":
		return Location{"Kastelorizo Island", "Europe/Athens", 36.141701, 29.576401}
	case "LAD":
		return Location{"Luanda", "Africa/Luanda", -8.858370, 13.231200}
	case "LAE":
		return Location{"Nadzab", "Pacific/Port_Moresby", -6.569830, 146.725998}
	case "LAH":
		return Location{"Labuha-Halmahera Island", "Asia/Jayapura", -0.635259, 127.501999}
	case "LAK":
		return Location{"Aklavik", "America/Inuvik", 68.223297, -135.005997}
	case "LAN":
		return Location{"Lansing", "America/Detroit", 42.778702, -84.587402}
	case "LAO":
		return Location{"Laoag City", "Asia/Manila", 18.178101, 120.531998}
	case "LAP":
		return Location{"La Paz", "America/Mazatlan", 24.072701, -110.362000}
	case "LAQ":
		return Location{"Al Bayda'", "Africa/Tripoli", 32.788700, 21.964300}
	case "LAR":
		return Location{"Laramie", "America/Denver", 41.312099, -105.675003}
	case "LAS":
		return Location{"Las Vegas", "America/Los_Angeles", 36.080101, -115.152000}
	case "LAU":
		return Location{"Lamu", "Africa/Nairobi", -2.252420, 40.913101}
	case "LAW":
		return Location{"Lawton", "America/Chicago", 34.567699, -98.416603}
	case "LAX":
		return Location{"Los Angeles", "America/Los_Angeles", 33.942501, -118.407997}
	case "LBA":
		return Location{"Leeds", "Europe/London", 53.865898, -1.660570}
	case "LBB":
		return Location{"Lubbock", "America/Chicago", 33.663601, -101.822998}
	case "LBD":
		return Location{"Khudzhand", "Asia/Dushanbe", 40.215401, 69.694702}
	case "LBE":
		return Location{"Latrobe", "America/New_York", 40.275902, -79.404800}
	case "LBF":
		return Location{"North Platte", "America/Chicago", 41.126202, -100.683998}
	case "LBJ":
		return Location{"Labuan Bajo-Flores Island", "Asia/Makassar", -8.486660, 119.889000}
	case "LBL":
		return Location{"Liberal", "America/Chicago", 37.044201, -100.959999}
	case "LBR":
		return Location{"Labrea", "America/Manaus", -7.278970, -64.769501}
	case "LBS":
		return Location{"", "Pacific/Fiji", -16.466700, 179.339996}
	case "LBU":
		return Location{"Labuan", "Asia/Kuching", 5.300680, 115.250000}
	case "LBV":
		return Location{"Libreville", "Africa/Libreville", 0.458600, 9.412280}
	case "LCA":
		return Location{"Larnarca", "Asia/Nicosia", 34.875099, 33.624901}
	case "LCE":
		return Location{"La Ceiba", "America/Tegucigalpa", 15.742500, -86.852997}
	case "LCG":
		return Location{"Culleredo", "Europe/Madrid", 43.302101, -8.377260}
	case "LCH":
		return Location{"Lake Charles", "America/Chicago", 30.126101, -93.223297}
	case "LCJ":
		return Location{"Lodz", "Europe/Warsaw", 51.721901, 19.398100}
	case "LCK":
		return Location{"Columbus", "America/New_York", 39.813801, -82.927803}
	case "LCX":
		return Location{"Longyan", "Asia/Shanghai", 25.674700, 116.747002}
	case "LCY":
		return Location{"London", "Europe/London", 51.505299, 0.055278}
	case "LDB":
		return Location{"Londrina", "America/Sao_Paulo", -23.333599, -51.130100}
	case "LDE":
		return Location{"Tarbes/Lourdes/Pyrenees", "Europe/Paris", 43.178699, -0.006439}
	case "LDH":
		return Location{"Lord Howe Island", "Australia/Lord_Howe", -31.538300, 159.076996}
	case "LDS":
		return Location{"Yichun", "Asia/Shanghai", 47.752056, 129.019125}
	case "LDU":
		return Location{"Lahad Datu", "Asia/Kuching", 5.032250, 118.323997}
	case "LDY":
		return Location{"Derry", "Europe/London", 55.042801, -7.161110}
	case "LEA":
		return Location{"Exmouth", "Australia/Perth", -22.235600, 114.088997}
	case "LEB":
		return Location{"Lebanon", "America/New_York", 43.626099, -72.304199}
	case "LEC":
		return Location{"Lencois", "America/Bahia", -12.482300, -41.277000}
	case "LED":
		return Location{"St. Petersburg", "Europe/Moscow", 59.800301, 30.262501}
	case "LEI":
		return Location{"Almeria", "Europe/Madrid", 36.843899, -2.370100}
	case "LEJ":
		return Location{"Leipzig", "Europe/Berlin", 51.432400, 12.241600}
	case "LEN":
		return Location{"Leon", "Europe/Madrid", 42.589001, -5.655560}
	case "LET":
		return Location{"Leticia", "America/Bogota", -4.193550, -69.943200}
	case "LEU":
		return Location{"Montferrer / Castellbo", "Europe/Madrid", 42.338600, 1.409170}
	case "LEX":
		return Location{"Lexington", "America/New_York", 38.036499, -84.605904}
	case "LFR":
		return Location{"", "America/Caracas", 8.239167, -72.271027}
	case "LFT":
		return Location{"Lafayette", "America/Chicago", 30.205299, -91.987602}
	case "LFW":
		return Location{"Lome", "Africa/Lome", 6.165610, 1.254510}
	case "LGA":
		return Location{"New York", "America/New_York", 40.777199, -73.872597}
	case "LGB":
		return Location{"Long Beach", "America/Los_Angeles", 33.817699, -118.152000}
	case "LGG":
		return Location{"Liege", "Europe/Brussels", 50.637402, 5.443220}
	case "LGI":
		return Location{"Deadman's Cay", "America/Nassau", 23.179001, -75.093597}
	case "LGK":
		return Location{"Langkawi", "Asia/Kuala_Lumpur", 6.329730, 99.728699}
	case "LGL":
		return Location{"Long Datih", "Asia/Kuching", 3.421000, 115.153999}
	case "LGW":
		return Location{"London", "Europe/London", 51.148102, -0.190278}
	case "LHE":
		return Location{"Lahore", "Asia/Karachi", 31.521601, 74.403603}
	case "LHG":
		return Location{"", "Australia/Sydney", -29.456699, 147.983994}
	case "LHR":
		return Location{"London", "Europe/London", 51.470600, -0.461941}
	case "LHW":
		return Location{"Lanzhou", "Asia/Shanghai", 36.515202, 103.620003}
	case "LIF":
		return Location{"Lifou", "Pacific/Noumea", -20.774799, 167.240005}
	case "LIG":
		return Location{"Limoges/Bellegarde", "Europe/Paris", 45.862801, 1.179440}
	case "LIH":
		return Location{"Lihue", "Pacific/Honolulu", 21.976000, -159.339005}
	case "LIL":
		return Location{"Lille/Lesquin", "Europe/Paris", 50.561901, 3.089440}
	case "LIM":
		return Location{"Lima", "America/Lima", -12.021900, -77.114304}
	case "LIN":
		return Location{"Milan", "Europe/Rome", 45.445099, 9.276740}
	case "LIO":
		return Location{"Puerto Limon", "America/Costa_Rica", 9.957960, -83.022003}
	case "LIR":
		return Location{"Liberia", "America/Costa_Rica", 10.593300, -85.544403}
	case "LIS":
		return Location{"Lisbon", "Europe/Lisbon", 38.781300, -9.135920}
	case "LIT":
		return Location{"Little Rock", "America/Chicago", 34.729401, -92.224297}
	case "LJG":
		return Location{"Lijiang", "Asia/Shanghai", 26.680000, 100.246002}
	case "LJU":
		return Location{"Ljubljana", "Europe/Ljubljana", 46.223701, 14.457600}
	case "LKA":
		return Location{"Larantuka-Flores Island", "Asia/Makassar", -8.346660, 122.981003}
	case "LKB":
		return Location{"Lakeba Island", "Pacific/Fiji", -18.199200, -178.817001}
	case "LKH":
		return Location{"Long Akah", "Asia/Kuching", 3.300000, 114.782997}
	case "LKL":
		return Location{"Lakselv", "Europe/Oslo", 70.068802, 24.973499}
	case "LKN":
		return Location{"Leknes", "Europe/Oslo", 68.152496, 13.609400}
	case "LKO":
		return Location{"Lucknow", "Asia/Kolkata", 26.760599, 80.889297}
	case "LKY":
		return Location{"Lake Manyara National Park", "Africa/Dar_es_Salaam", -3.376310, 35.818298}
	case "LLA":
		return Location{"Lulea", "Europe/Stockholm", 65.543800, 22.122000}
	case "LLF":
		return Location{"Yongzhou", "Asia/Shanghai", 26.338661, 111.610043}
	case "LLI":
		return Location{"Lalibela", "Africa/Addis_Ababa", 11.975000, 38.980000}
	case "LLJ":
		return Location{"Lalmonirhat", "Asia/Dhaka", 25.887501, 89.433098}
	case "LLK":
		return Location{"Lankaran", "Asia/Baku", 38.746399, 48.818001}
	case "LLW":
		return Location{"Lilongwe", "Africa/Blantyre", -13.789400, 33.780998}
	case "LMA":
		return Location{"Minchumina", "America/Anchorage", 63.886002, -152.302002}
	case "LMM":
		return Location{"Los Mochis", "America/Mazatlan", 25.685200, -109.081001}
	case "LMN":
		return Location{"Limbang", "Asia/Brunei", 4.808300, 115.010002}
	case "LMP":
		return Location{"Lampedusa", "Europe/Rome", 35.497898, 12.618100}
	case "LNB":
		return Location{"Lamen Bay", "Pacific/Efate", -16.584200, 168.158997}
	case "LNE":
		return Location{"Lonorore", "Pacific/Efate", -15.865600, 168.171997}
	case "LNJ":
		return Location{"Lincang", "Asia/Shanghai", 23.738333, 100.025000}
	case "LNK":
		return Location{"Lincoln", "America/Chicago", 40.851002, -96.759201}
	case "LNO":
		return Location{"Leonora", "Australia/Perth", -28.878099, 121.315002}
	case "LNS":
		return Location{"Lancaster", "America/New_York", 40.121700, -76.296097}
	case "LNV":
		return Location{"Londolovit", "Pacific/Port_Moresby", -3.043610, 152.628998}
	case "LNY":
		return Location{"Lanai City", "Pacific/Honolulu", 20.785601, -156.951004}
	case "LNZ":
		return Location{"Linz", "Europe/Vienna", 48.233200, 14.187500}
	case "LOD":
		return Location{"Longana", "Pacific/Efate", -15.306700, 167.966995}
	case "LOE":
		return Location{"", "Asia/Bangkok", 17.439100, 101.722000}
	case "LOH":
		return Location{"La Toma (Catamayo)", "America/Guayaquil", -3.995890, -79.371902}
	case "LOK":
		return Location{"Lodwar", "Africa/Nairobi", 3.121970, 35.608700}
	case "LOO":
		return Location{"Laghouat", "Africa/Algiers", 33.764400, 2.928340}
	case "LOP":
		return Location{"Mataram", "Asia/Makassar", -8.757322, 116.276675}
	case "LOS":
		return Location{"Lagos", "Africa/Lagos", 6.577370, 3.321160}
	case "LOY":
		return Location{"Loyengalani", "Africa/Nairobi", 2.750000, 36.716999}
	case "LPA":
		return Location{"Gran Canaria Island", "Atlantic/Canary", 27.931900, -15.386600}
	case "LPB":
		return Location{"La Paz / El Alto", "America/La_Paz", -16.513300, -68.192299}
	case "LPI":
		return Location{"Linkoping", "Europe/Stockholm", 58.406200, 15.680500}
	case "LPL":
		return Location{"Liverpool", "Europe/London", 53.333599, -2.849720}
	case "LPM":
		return Location{"Lamap", "Pacific/Efate", -16.454000, 167.822998}
	case "LPP":
		return Location{"Lappeenranta", "Europe/Helsinki", 61.044601, 28.144400}
	case "LPQ":
		return Location{"Luang Phabang", "Asia/Vientiane", 19.897301, 102.161003}
	case "LPT":
		return Location{"", "Asia/Bangkok", 18.270901, 99.504204}
	case "LPY":
		return Location{"Le Puy/Loudes", "Europe/Paris", 45.080700, 3.762890}
	case "LQM":
		return Location{"Puerto Leguizamo", "America/Bogota", -0.182278, -74.770800}
	case "LRD":
		return Location{"Laredo", "America/Chicago", 27.543800, -99.461601}
	case "LRE":
		return Location{"Longreach", "Australia/Brisbane", -23.434200, 144.279999}
	case "LRH":
		return Location{"La Rochelle/Ile de Re", "Europe/Paris", 46.179199, -1.195280}
	case "LRM":
		return Location{"La Romana", "America/Santo_Domingo", 18.450701, -68.911797}
	case "LRR":
		return Location{"Lar", "Asia/Tehran", 27.674700, 54.383301}
	case "LRS":
		return Location{"Leros Island", "Europe/Athens", 37.184898, 26.800301}
	case "LRT":
		return Location{"Lorient/Lann/Bihoue", "Europe/Paris", 47.760601, -3.440000}
	case "LRU":
		return Location{"Las Cruces", "America/Denver", 32.289398, -106.921997}
	case "LRV":
		return Location{"Los Roques", "America/Caracas", 11.950000, -66.669998}
	case "LSC":
		return Location{"La Serena-Coquimbo", "America/Santiago", -29.916201, -71.199501}
	case "LSE":
		return Location{"La Crosse", "America/Chicago", 43.879002, -91.256699}
	case "LSH":
		return Location{"Lashio", "Asia/Yangon", 22.977900, 97.752197}
	case "LSI":
		return Location{"Lerwick", "Europe/London", 59.878899, -1.295560}
	case "LSP":
		return Location{"Paraguana", "America/Caracas", 11.780775, -70.151497}
	case "LSQ":
		return Location{"Los Angeles", "America/Santiago", -37.401699, -72.425400}
	case "LST":
		return Location{"Launceston", "Australia/Hobart", -41.545300, 147.214005}
	case "LSW":
		return Location{"Lhok Seumawe-Sumatra Island", "Asia/Jakarta", 5.226680, 96.950302}
	case "LTI":
		return Location{"Altai", "Asia/Hovd", 46.376400, 96.221100}
	case "LTN":
		return Location{"London", "Europe/London", 51.874699, -0.368333}
	case "LTO":
		return Location{"Loreto", "America/Mazatlan", 25.989201, -111.348000}
	case "LUD":
		return Location{"Luderitz", "Africa/Windhoek", -26.687401, 15.242900}
	case "LUM":
		return Location{"Luxi", "Asia/Shanghai", 24.401100, 98.531700}
	case "LUN":
		return Location{"Lusaka", "Africa/Lusaka", -15.330800, 28.452600}
	case "LUO":
		return Location{"Luena", "Africa/Luanda", -11.768100, 19.897699}
	case "LUP":
		return Location{"Kalaupapa", "Pacific/Honolulu", 21.211000, -156.973999}
	case "LUQ":
		return Location{"San Luis", "America/Argentina/San_Luis", -33.273201, -66.356400}
	case "LUR":
		return Location{"Cape Lisburne", "America/Nome", 68.875099, -166.110001}
	case "LUV":
		return Location{"Langgur-Seram Island", "Asia/Jayapura", -5.661620, 132.731003}
	case "LUW":
		return Location{"Luwok-Celebes Island", "Asia/Makassar", -1.038920, 122.772003}
	case "LUX":
		return Location{"Luxembourg", "Europe/Luxembourg", 49.626598, 6.211520}
	case "LUZ":
		return Location{"Lublin", "Europe/Warsaw", 51.240278, 22.713611}
	case "LVI":
		return Location{"Livingstone", "Africa/Lusaka", -17.821800, 25.822701}
	case "LVO":
		return Location{"", "Australia/Perth", -28.613600, 122.424004}
	case "LWB":
		return Location{"Lewisburg", "America/New_York", 37.858299, -80.399498}
	case "LWS":
		return Location{"Lewiston", "America/Los_Angeles", 46.374500, -117.014999}
	case "LWY":
		return Location{"Lawas", "Asia/Kuching", 4.849170, 115.407997}
	case "LXA":
		return Location{"Lhasa", "Asia/Shanghai", 29.297800, 90.911903}
	case "LXG":
		return Location{"Luang Namtha", "Asia/Vientiane", 20.966999, 101.400002}
	case "LXR":
		return Location{"Luxor", "Africa/Cairo", 25.671000, 32.706600}
	case "LXS":
		return Location{"Limnos Island", "Europe/Athens", 39.917099, 25.236300}
	case "LYA":
		return Location{"Luoyang", "Asia/Shanghai", 34.741100, 112.388000}
	case "LYB":
		return Location{"Little Cayman", "America/Cayman", 19.667000, -80.099998}
	case "LYC":
		return Location{"", "Europe/Stockholm", 64.548302, 18.716200}
	case "LYG":
		return Location{"Lianyungang", "Asia/Shanghai", 34.571667, 118.873611}
	case "LYH":
		return Location{"Lynchburg", "America/New_York", 37.326698, -79.200401}
	case "LYI":
		return Location{"Linyi", "Asia/Shanghai", 35.046101, 118.412003}
	case "LYP":
		return Location{"Faisalabad", "Asia/Karachi", 31.365000, 72.994797}
	case "LYR":
		return Location{"Longyearbyen", "Arctic/Longyearbyen", 78.246101, 15.465600}
	case "LYS":
		return Location{"Lyon", "Europe/Paris", 45.726398, 5.090830}
	case "LZH":
		return Location{"Liuzhou", "Asia/Shanghai", 24.207500, 109.390999}
	case "LZN":
		return Location{"Nangang Island", "Asia/Taipei", 26.159800, 119.958000}
	case "LZO":
		return Location{"Luzhou", "Asia/Shanghai", 28.852200, 105.392998}
	case "LZY":
		return Location{"Nyingchi", "Asia/Shanghai", 29.303301, 94.335297}
	case "MAA":
		return Location{"Chennai", "Asia/Kolkata", 12.990005, 80.169296}
	case "MAB":
		return Location{"Maraba", "America/Belem", -5.368590, -49.138000}
	case "MAD":
		return Location{"Madrid", "Europe/Madrid", 40.493600, -3.566760}
	case "MAF":
		return Location{"Midland", "America/Chicago", 31.942499, -102.202003}
	case "MAG":
		return Location{"Madang", "Pacific/Port_Moresby", -5.207080, 145.789001}
	case "MAH":
		return Location{"Menorca Island", "Europe/Madrid", 39.862598, 4.218650}
	case "MAJ":
		return Location{"Majuro Atoll", "Pacific/Majuro", 7.064760, 171.272003}
	case "MAM":
		return Location{"Matamoros", "America/Matamoros", 25.769899, -97.525299}
	case "MAN":
		return Location{"Manchester", "Europe/London", 53.353699, -2.274950}
	case "MAO":
		return Location{"Manaus", "America/Manaus", -3.038610, -60.049702}
	case "MAQ":
		return Location{"", "Asia/Bangkok", 16.699900, 98.545097}
	case "MAR":
		return Location{"Maracaibo", "America/Caracas", 10.558208, -71.727859}
	case "MAS":
		return Location{"", "Pacific/Port_Moresby", -2.061890, 147.423996}
	case "MAU":
		return Location{"", "Pacific/Tahiti", -16.426500, -152.244003}
	case "MAZ":
		return Location{"Mayaguez", "America/Puerto_Rico", 18.255699, -67.148499}
	case "MBA":
		return Location{"Mombasa", "Africa/Nairobi", -4.034830, 39.594200}
	case "MBE":
		return Location{"Monbetsu", "Asia/Tokyo", 44.303902, 143.404007}
	case "MBI":
		return Location{"Mbeya", "Africa/Dar_es_Salaam", -8.917000, 33.466999}
	case "MBJ":
		return Location{"Montego Bay", "America/Jamaica", 18.503700, -77.913399}
	case "MBL":
		return Location{"Manistee", "America/Detroit", 44.272400, -86.246902}
	case "MBQ":
		return Location{"Mbarara", "Africa/Kampala", -0.555278, 30.599400}
	case "MBS":
		return Location{"Saginaw", "America/Detroit", 43.532902, -84.079597}
	case "MBT":
		return Location{"Masbate", "Asia/Manila", 12.369400, 123.628998}
	case "MBZ":
		return Location{"Maues", "America/Manaus", -3.372170, -57.724800}
	case "MCE":
		return Location{"Merced", "America/Los_Angeles", 37.284698, -120.514000}
	case "MCI":
		return Location{"Kansas City", "America/Chicago", 39.297600, -94.713898}
	case "MCK":
		return Location{"Mc Cook", "America/Chicago", 40.206299, -100.592003}
	case "MCN":
		return Location{"Macon", "America/New_York", 32.692799, -83.649200}
	case "MCO":
		return Location{"Orlando", "America/New_York", 28.429399, -81.308998}
	case "MCP":
		return Location{"Macapa", "America/Belem", 0.050664, -51.072201}
	case "MCT":
		return Location{"Muscat", "Asia/Muscat", 23.593300, 58.284401}
	case "MCV":
		return Location{"McArthur River Mine", "Australia/Darwin", -16.442499, 136.084000}
	case "MCW":
		return Location{"Mason City", "America/Chicago", 43.157799, -93.331299}
	case "MCX":
		return Location{"Makhachkala", "Europe/Moscow", 42.816799, 47.652302}
	case "MCY":
		return Location{"Maroochydore", "Australia/Brisbane", -26.603300, 153.091003}
	case "MCZ":
		return Location{"Maceio", "America/Maceio", -9.510810, -35.791698}
	case "MDC":
		return Location{"Manado-Celebes Island", "Asia/Makassar", 1.549260, 124.926003}
	case "MDE":
		return Location{"Rionegro", "America/Bogota", 6.164540, -75.423100}
	case "MDG":
		return Location{"Mudanjiang", "Asia/Shanghai", 44.524101, 129.569000}
	case "MDI":
		return Location{"Makurdi", "Africa/Lagos", 7.703880, 8.613940}
	case "MDK":
		return Location{"Mbandaka", "Africa/Kinshasa", 0.022600, 18.288700}
	case "MDL":
		return Location{"Mandalay", "Asia/Yangon", 21.702200, 95.977898}
	case "MDQ":
		return Location{"Mar del Plata", "America/Argentina/Buenos_Aires", -37.934200, -57.573300}
	case "MDT":
		return Location{"Harrisburg", "America/New_York", 40.193501, -76.763397}
	case "MDW":
		return Location{"Chicago", "America/Chicago", 41.785999, -87.752403}
	case "MDZ":
		return Location{"Mendoza", "America/Argentina/Mendoza", -32.831699, -68.792900}
	case "MEB":
		return Location{"", "Australia/Melbourne", -37.728100, 144.901993}
	case "MEC":
		return Location{"Manta", "America/Guayaquil", -0.946078, -80.678802}
	case "MED":
		return Location{"Medina", "Asia/Riyadh", 24.553400, 39.705101}
	case "MEE":
		return Location{"Mare", "Pacific/Noumea", -21.481701, 168.037994}
	case "MEH":
		return Location{"Mehamn", "Europe/Oslo", 71.029701, 27.826700}
	case "MEI":
		return Location{"Meridian", "America/Chicago", 32.332600, -88.751900}
	case "MEL":
		return Location{"Melbourne", "Australia/Melbourne", -37.673302, 144.843002}
	case "MEM":
		return Location{"Memphis", "America/Chicago", 35.042400, -89.976700}
	case "MEQ":
		return Location{"Peureumeue-Sumatra Island", "Asia/Jakarta", 4.250000, 96.217003}
	case "MEU":
		return Location{"Almeirim", "America/Santarem", -0.889839, -52.602200}
	case "MEX":
		return Location{"Mexico City", "America/Mexico_City", 19.436300, -99.072098}
	case "MFA":
		return Location{"Mafia Island", "Africa/Dar_es_Salaam", -7.917000, 39.667000}
	case "MFE":
		return Location{"Mc Allen", "America/Chicago", 26.175800, -98.238602}
	case "MFK":
		return Location{"Beigan Island", "Asia/Taipei", 26.224199, 120.002998}
	case "MFM":
		return Location{"Taipa", "Asia/Macau", 22.149599, 113.592003}
	case "MFR":
		return Location{"Medford", "America/Los_Angeles", 42.374199, -122.873001}
	case "MFU":
		return Location{"Mfuwe", "Africa/Lusaka", -13.258900, 31.936600}
	case "MGA":
		return Location{"Managua", "America/Managua", 12.141500, -86.168198}
	case "MGB":
		return Location{"", "Australia/Adelaide", -37.745602, 140.785004}
	case "MGF":
		return Location{"Maringa", "America/Sao_Paulo", -23.479445, -52.012222}
	case "MGH":
		return Location{"Margate", "Africa/Johannesburg", -30.857401, 30.343000}
	case "MGM":
		return Location{"Montgomery", "America/Chicago", 32.300598, -86.393997}
	case "MGQ":
		return Location{"Mogadishu", "Africa/Mogadishu", 2.014440, 45.304699}
	case "MGT":
		return Location{"Milingimbi Island", "Australia/Darwin", -12.094400, 134.893997}
	case "MGW":
		return Location{"Morgantown", "America/New_York", 39.642899, -79.916298}
	case "MGZ":
		return Location{"Mkeik", "Asia/Yangon", 12.439800, 98.621498}
	case "MHC":
		return Location{"Dalcahue", "America/Santiago", -42.340278, -73.715556}
	case "MHD":
		return Location{"Mashhad", "Asia/Tehran", 36.235199, 59.640999}
	case "MHH":
		return Location{"Marsh Harbour", "America/Nassau", 26.511400, -77.083504}
	case "MHK":
		return Location{"Manhattan", "America/Chicago", 39.140999, -96.670799}
	case "MHQ":
		return Location{"", "Europe/Mariehamn", 60.122200, 19.898199}
	case "MHT":
		return Location{"Manchester", "America/New_York", 42.932598, -71.435699}
	case "MIA":
		return Location{"Miami", "America/New_York", 25.793200, -80.290604}
	case "MID":
		return Location{"Merida", "America/Merida", 20.937000, -89.657700}
	case "MIG":
		return Location{"Mianyang", "Asia/Shanghai", 31.428101, 104.740997}
	case "MII":
		return Location{"Marilia", "America/Sao_Paulo", -22.196899, -49.926399}
	case "MIM":
		return Location{"Merimbula", "Australia/Sydney", -36.908600, 149.901001}
	case "MIR":
		return Location{"Monastir", "Africa/Tunis", 35.758099, 10.754700}
	case "MIS":
		return Location{"Misima Island", "Pacific/Port_Moresby", -10.689200, 152.837997}
	case "MIU":
		return Location{"Maiduguri", "Africa/Lagos", 11.855300, 13.080900}
	case "MJC":
		return Location{"", "Africa/Abidjan", 7.272070, -7.587360}
	case "MJF":
		return Location{"", "Europe/Oslo", 65.783997, 13.214900}
	case "MJI":
		return Location{"Tripoli", "Africa/Tripoli", 32.894100, 13.276000}
	case "MJK":
		return Location{"Monkey Mia", "Australia/Perth", -25.893900, 113.577003}
	case "MJM":
		return Location{"Mbuji Mayi", "Africa/Lubumbashi", -6.121240, 23.569000}
	case "MJN":
		return Location{"", "Indian/Antananarivo", -15.666842, 46.351233}
	case "MJT":
		return Location{"Mytilene", "Europe/Athens", 39.056702, 26.598301}
	case "MJU":
		return Location{"Mamuju-Celebes Island", "Asia/Makassar", -2.583333, 119.033333}
	case "MJZ":
		return Location{"Mirny", "Asia/Yakutsk", 62.534698, 114.039001}
	case "MKE":
		return Location{"Milwaukee", "America/Chicago", 42.947201, -87.896599}
	case "MKG":
		return Location{"Muskegon", "America/Detroit", 43.169498, -86.238197}
	case "MKK":
		return Location{"Kaunakakai", "Pacific/Honolulu", 21.152901, -157.095993}
	case "MKL":
		return Location{"Jackson", "America/Chicago", 35.599899, -88.915604}
	case "MKM":
		return Location{"Mukah", "Asia/Kuching", 2.906390, 112.080002}
	case "MKP":
		return Location{"", "Pacific/Tahiti", -16.583900, -143.658005}
	case "MKQ":
		return Location{"Merauke-Papua Island", "Asia/Jayapura", -8.520290, 140.417999}
	case "MKR":
		return Location{"", "Australia/Perth", -26.611700, 118.547997}
	case "MKW":
		return Location{"Manokwari-Papua Island", "Asia/Jayapura", -0.891833, 134.048996}
	case "MKY":
		return Location{"Mackay", "Australia/Brisbane", -21.171700, 149.179993}
	case "MLA":
		return Location{"Luqa", "Europe/Malta", 35.857498, 14.477500}
	case "MLB":
		return Location{"Melbourne", "America/New_York", 28.102800, -80.645302}
	case "MLE":
		return Location{"Male", "Indian/Maldives", 4.191830, 73.529099}
	case "MLG":
		return Location{"Malang-Java Island", "Asia/Jakarta", -7.926560, 112.714996}
	case "MLH":
		return Location{"Bale/Mulhouse", "Europe/Paris", 47.589600, 7.529910}
	case "MLI":
		return Location{"Moline", "America/Chicago", 41.448502, -90.507500}
	case "MLL":
		return Location{"Marshall", "America/Anchorage", 61.864300, -162.026001}
	case "MLM":
		return Location{"Morelia", "America/Mexico_City", 19.849899, -101.025002}
	case "MLN":
		return Location{"Melilla Island", "Africa/Casablanca", 35.279800, -2.956260}
	case "MLO":
		return Location{"Milos Island", "Europe/Athens", 36.696899, 24.476900}
	case "MLU":
		return Location{"Monroe", "America/Chicago", 32.510899, -92.037697}
	case "MLX":
		return Location{"Malatya", "Europe/Istanbul", 38.435299, 38.091000}
	case "MLY":
		return Location{"Manley Hot Springs", "America/Anchorage", 64.997597, -150.643997}
	case "MMB":
		return Location{"Ozora", "Asia/Tokyo", 43.880600, 144.164001}
	case "MME":
		return Location{"Durham", "Europe/London", 54.509201, -1.429410}
	case "MMG":
		return Location{"", "Australia/Perth", -28.116100, 117.842003}
	case "MMH":
		return Location{"Mammoth Lakes", "America/Los_Angeles", 37.624100, -118.837997}
	case "MMJ":
		return Location{"Matsumoto", "Asia/Tokyo", 36.166801, 137.923004}
	case "MMK":
		return Location{"Murmansk", "Europe/Moscow", 68.781700, 32.750801}
	case "MMO":
		return Location{"Vila do Maio", "Atlantic/Cape_Verde", 15.155900, -23.213699}
	case "MMU":
		return Location{"Morristown", "America/New_York", 40.799400, -74.414902}
	case "MMX":
		return Location{"Malmo", "Europe/Stockholm", 55.536305, 13.376198}
	case "MMY":
		return Location{"Miyako City", "Asia/Tokyo", 24.782801, 125.294998}
	case "MNA":
		return Location{"Karakelong Island", "Asia/Makassar", 4.006940, 126.672997}
	case "MNC":
		return Location{"Nacala", "Africa/Maputo", -14.488200, 40.712200}
	case "MNG":
		return Location{"Maningrida", "Australia/Darwin", -12.056100, 134.233994}
	case "MNI":
		return Location{"Gerald's Park", "America/Montserrat", 16.791401, -62.193298}
	case "MNL":
		return Location{"Manila", "Asia/Manila", 14.508600, 121.019997}
	case "MNS":
		return Location{"Mansa", "Africa/Lusaka", -11.137000, 28.872601}
	case "MNX":
		return Location{"Manicore", "America/Manaus", -5.811380, -61.278301}
	case "MOB":
		return Location{"Mobile", "America/Chicago", 30.691200, -88.242798}
	case "MOC":
		return Location{"Montes Claros", "America/Sao_Paulo", -16.706900, -43.818901}
	case "MOF":
		return Location{"Maumere-Flores Island", "Asia/Makassar", -8.640650, 122.237000}
	case "MOL":
		return Location{"Aro", "Europe/Oslo", 62.744701, 7.262500}
	case "MOQ":
		return Location{"", "Indian/Antananarivo", -20.284700, 44.317600}
	case "MOT":
		return Location{"Minot", "America/Chicago", 48.259399, -101.279999}
	case "MOU":
		return Location{"Mountain Village", "America/Nome", 62.095402, -163.682007}
	case "MOV":
		return Location{"Moranbah", "Australia/Brisbane", -22.057800, 148.076996}
	case "MOZ":
		return Location{"", "Pacific/Tahiti", -17.490000, -149.761993}
	case "MPA":
		return Location{"Mpacha", "Africa/Windhoek", -17.634399, 24.176701}
	case "MPH":
		return Location{"Malay", "Asia/Manila", 11.924500, 121.954002}
	case "MPL":
		return Location{"Montpellier/Mediterranee", "Europe/Paris", 43.576199, 3.963010}
	case "MPM":
		return Location{"Maputo", "Africa/Maputo", -25.920799, 32.572601}
	case "MPN":
		return Location{"Mount Pleasant", "Atlantic/Stanley", -51.822800, -58.447201}
	case "MQF":
		return Location{"Magnitogorsk", "Asia/Yekaterinburg", 53.393101, 58.755699}
	case "MQJ":
		return Location{"Honuu", "Asia/Srednekolymsk", 66.450859, 143.261551}
	case "MQL":
		return Location{"Mildura", "Australia/Melbourne", -34.229198, 142.085999}
	case "MQM":
		return Location{"Mardin", "Europe/Istanbul", 37.223301, 40.631699}
	case "MQN":
		return Location{"Mo i Rana", "Europe/Oslo", 66.363899, 14.301400}
	case "MQP":
		return Location{"Mpumalanga", "Africa/Johannesburg", -25.383200, 31.105600}
	case "MQT":
		return Location{"Marquette", "America/Detroit", 46.353600, -87.395401}
	case "MQX":
		return Location{"", "Africa/Addis_Ababa", 13.467400, 39.533501}
	case "MRD":
		return Location{"Merida", "America/Caracas", 8.582078, -71.161041}
	case "MRE":
		return Location{"Masai Mara", "Africa/Nairobi", -1.406111, 35.008057}
	case "MRS":
		return Location{"Marseille", "Europe/Paris", 43.439272, 5.221424}
	case "MRU":
		return Location{"Port Louis", "Indian/Mauritius", -20.430201, 57.683601}
	case "MRV":
		return Location{"Mineralnyye Vody", "Europe/Moscow", 44.225101, 43.081902}
	case "MRX":
		return Location{"", "Asia/Tehran", 30.556200, 49.151901}
	case "MRY":
		return Location{"Monterey", "America/Los_Angeles", 36.587002, -121.843002}
	case "MRZ":
		return Location{"Moree", "Australia/Sydney", -29.498899, 149.845001}
	case "MSA":
		return Location{"Muskrat Dam", "America/Rainy_River", 53.441399, -91.762802}
	case "MSH":
		return Location{"Masirah", "Asia/Muscat", 20.675400, 58.890499}
	case "MSJ":
		return Location{"Misawa", "Asia/Tokyo", 40.703201, 141.367996}
	case "MSL":
		return Location{"Muscle Shoals", "America/Chicago", 34.745300, -87.610199}
	case "MSN":
		return Location{"Madison", "America/Chicago", 43.139900, -89.337502}
	case "MSO":
		return Location{"Missoula", "America/Denver", 46.916302, -114.091003}
	case "MSP":
		return Location{"Minneapolis", "America/Chicago", 44.882000, -93.221802}
	case "MSQ":
		return Location{"Minsk", "Europe/Minsk", 53.882500, 28.030701}
	case "MSR":
		return Location{"Mus", "Europe/Istanbul", 38.747799, 41.661201}
	case "MSS":
		return Location{"Massena", "America/New_York", 44.935799, -74.845596}
	case "MST":
		return Location{"Maastricht", "Europe/Amsterdam", 50.911701, 5.770140}
	case "MSU":
		return Location{"Maseru", "Africa/Maseru", -29.462299, 27.552500}
	case "MSY":
		return Location{"New Orleans", "America/Chicago", 29.993401, -90.258003}
	case "MSZ":
		return Location{"Namibe", "Africa/Luanda", -15.261200, 12.146800}
	case "MTE":
		return Location{"Monte Alegre", "America/Santarem", -1.995800, -54.074200}
	case "MTJ":
		return Location{"Montrose", "America/Denver", 38.509800, -107.893997}
	case "MTR":
		return Location{"Monteria", "America/Bogota", 8.823740, -75.825800}
	case "MTT":
		return Location{"Minatitlan", "America/Mexico_City", 18.103399, -94.580704}
	case "MTV":
		return Location{"Ablow", "Pacific/Efate", -13.666000, 167.712006}
	case "MTY":
		return Location{"Monterrey", "America/Monterrey", 25.778500, -100.107002}
	case "MUA":
		return Location{"", "Pacific/Guadalcanal", -8.327970, 157.263000}
	case "MUB":
		return Location{"Maun", "Africa/Gaborone", -19.972601, 23.431101}
	case "MUC":
		return Location{"Munich", "Europe/Berlin", 48.353802, 11.786100}
	case "MUE":
		return Location{"Kamuela", "Pacific/Honolulu", 20.001301, -155.667999}
	case "MUH":
		return Location{"Mersa Matruh", "Africa/Cairo", 31.325399, 27.221701}
	case "MUN":
		return Location{"", "America/Caracas", 9.754530, -63.147400}
	case "MUR":
		return Location{"Marudi", "Asia/Kuching", 4.178980, 114.329002}
	case "MUX":
		return Location{"Multan", "Asia/Karachi", 30.203199, 71.419098}
	case "MVB":
		return Location{"Franceville", "Africa/Libreville", -1.656160, 13.438000}
	case "MVD":
		return Location{"Montevideo", "America/Montevideo", -34.838402, -56.030800}
	case "MVF":
		return Location{"Mossoro", "America/Fortaleza", -5.201920, -37.364300}
	case "MVP":
		return Location{"Mitu", "America/Bogota", 1.253660, -70.233900}
	case "MVR":
		return Location{"Maroua", "Africa/Douala", 10.451400, 14.257400}
	case "MVT":
		return Location{"", "Pacific/Tahiti", -14.868100, -148.716995}
	case "MVY":
		return Location{"Martha's Vineyard", "America/New_York", 41.393101, -70.614304}
	case "MWA":
		return Location{"Marion", "America/Chicago", 37.755001, -89.011101}
	case "MWF":
		return Location{"Maewo Island", "Pacific/Efate", -15.000000, 168.082993}
	case "MWX":
		return Location{"", "Asia/Seoul", 34.991406, 126.382814}
	case "MWZ":
		return Location{"Mwanza", "Africa/Dar_es_Salaam", -2.444490, 32.932701}
	case "MXH":
		return Location{"Moro", "Pacific/Port_Moresby", -6.363330, 143.238007}
	case "MXL":
		return Location{"Mexicali", "America/Tijuana", 32.630600, -115.241997}
	case "MXP":
		return Location{"Milan", "Europe/Rome", 45.630600, 8.728110}
	case "MXV":
		return Location{"Moron", "Asia/Ulaanbaatar", 49.663300, 100.098999}
	case "MXX":
		return Location{"", "Europe/Stockholm", 60.957901, 14.511400}
	case "MXZ":
		return Location{"Meixian", "Asia/Shanghai", 24.350000, 116.133003}
	case "MYA":
		return Location{"Moruya", "Australia/Sydney", -35.897800, 150.143997}
	case "MYD":
		return Location{"Malindi", "Africa/Nairobi", -3.229310, 40.101700}
	case "MYG":
		return Location{"Mayaguana", "America/Nassau", 22.379499, -73.013496}
	case "MYI":
		return Location{"Murray Island", "Australia/Brisbane", -9.916670, 144.054993}
	case "MYJ":
		return Location{"Matsuyama", "Asia/Tokyo", 33.827202, 132.699997}
	case "MYP":
		return Location{"Mary", "Asia/Ashgabat", 37.619400, 61.896702}
	case "MYQ":
		return Location{"Mysore", "Asia/Kolkata", 12.307200, 76.649696}
	case "MYR":
		return Location{"Myrtle Beach", "America/New_York", 33.679699, -78.928299}
	case "MYT":
		return Location{"Myitkyina", "Asia/Yangon", 25.383600, 97.351898}
	case "MYU":
		return Location{"Mekoryuk", "America/Nome", 60.371399, -166.270996}
	case "MYW":
		return Location{"Mtwara", "Africa/Dar_es_Salaam", -10.339100, 40.181801}
	case "MYY":
		return Location{"Miri", "Asia/Kuching", 4.322010, 113.987000}
	case "MZA":
		return Location{"Mazamari", "America/Lima", -11.325400, -74.535599}
	case "MZG":
		return Location{"Makung City", "Asia/Taipei", 23.568701, 119.627998}
	case "MZH":
		return Location{"Amasya", "Europe/Istanbul", 40.829399, 35.521999}
	case "MZL":
		return Location{"Manizales", "America/Bogota", 5.029600, -75.464700}
	case "MZO":
		return Location{"Manzanillo", "America/Havana", 20.288099, -77.089203}
	case "MZR":
		return Location{"", "Asia/Kabul", 36.706902, 67.209702}
	case "MZT":
		return Location{"Mazatlan", "America/Mazatlan", 23.161400, -106.265999}
	case "MZV":
		return Location{"Mulu", "Asia/Kuching", 4.048330, 114.805000}
	case "MZW":
		return Location{"Mecheria", "Africa/Algiers", 33.535900, -0.242353}
	case "NAA":
		return Location{"Narrabri", "Australia/Sydney", -30.319201, 149.826996}
	case "NAG":
		return Location{"Naqpur", "Asia/Kolkata", 21.092199, 79.047203}
	case "NAH":
		return Location{"Tahuna-Sangihe Island", "Asia/Makassar", 3.683210, 125.528000}
	case "NAJ":
		return Location{"Nakhchivan", "Asia/Baku", 39.188801, 45.458401}
	case "NAL":
		return Location{"Nalchik", "Europe/Moscow", 43.512901, 43.636600}
	case "NAM":
		return Location{"Namlea-Buru Island", "Asia/Jayapura", -3.235570, 127.099998}
	case "NAN":
		return Location{"Nadi", "Pacific/Fiji", -17.755400, 177.442993}
	case "NAO":
		return Location{"Nanchong", "Asia/Shanghai", 30.795450, 106.162600}
	case "NAP":
		return Location{"Napoli", "Europe/Rome", 40.886002, 14.290800}
	case "NAQ":
		return Location{"Qaanaaq", "America/Thule", 77.488602, -69.388702}
	case "NAS":
		return Location{"Nassau", "America/Nassau", 25.039000, -77.466202}
	case "NAT":
		return Location{"Natal", "America/Fortaleza", -5.911420, -35.247700}
	case "NAU":
		return Location{"Napuka Island", "Pacific/Tahiti", -14.176800, -141.266998}
	case "NAV":
		return Location{"Nevsehir", "Europe/Istanbul", 38.771900, 34.534500}
	case "NAW":
		return Location{"", "Asia/Bangkok", 6.519920, 101.742996}
	case "NBC":
		return Location{"Nizhnekamsk", "Europe/Moscow", 55.564701, 52.092499}
	case "NBE":
		return Location{"Enfidha", "Africa/Tunis", 36.075833, 10.438611}
	case "NBO":
		return Location{"Nairobi", "Africa/Nairobi", -1.319240, 36.927799}
	case "NBX":
		return Location{"Nabire-Papua Island", "Asia/Jayapura", -3.368180, 135.496002}
	case "NCE":
		return Location{"Nice", "Europe/Paris", 43.658401, 7.215870}
	case "NCL":
		return Location{"Newcastle", "Europe/London", 55.037498, -1.691670}
	case "NCU":
		return Location{"Nukus", "Asia/Samarkand", 42.488400, 59.623299}
	case "NDB":
		return Location{"Nouadhibou", "Africa/El_Aaiun", 20.933100, -17.030001}
	case "NDG":
		return Location{"Qiqihar", "Asia/Shanghai", 47.239601, 123.917999}
	case "NDJ":
		return Location{"N'Djamena", "Africa/Ndjamena", 12.133700, 15.034000}
	case "NDR":
		return Location{"Nador", "Africa/Casablanca", 34.988800, -3.028210}
	case "NDU":
		return Location{"Rundu", "Africa/Windhoek", -17.956499, 19.719400}
	case "NEU":
		return Location{"", "Asia/Vientiane", 20.418400, 104.067001}
	case "NEV":
		return Location{"Charlestown", "America/St_Kitts", 17.205700, -62.589901}
	case "NGB":
		return Location{"Ningbo", "Asia/Shanghai", 29.826700, 121.461998}
	case "NGE":
		return Location{"N'Gaoundere", "Africa/Douala", 7.357010, 13.559200}
	case "NGO":
		return Location{"Tokoname", "Asia/Tokyo", 34.858398, 136.804993}
	case "NGQ":
		return Location{"Shiquanhe", "Asia/Shanghai", 32.100000, 80.053056}
	case "NGS":
		return Location{"Nagasaki", "Asia/Tokyo", 32.916901, 129.914001}
	case "NHV":
		return Location{"", "Pacific/Marquesas", -8.795600, -140.229004}
	case "NIM":
		return Location{"Niamey", "Africa/Niamey", 13.481500, 2.183610}
	case "NJC":
		return Location{"Nizhnevartovsk", "Asia/Yekaterinburg", 60.949299, 76.483597}
	case "NJF":
		return Location{"Najaf", "Asia/Baghdad", 31.989722, 44.404167}
	case "NKC":
		return Location{"Nouakchott", "Africa/Nouakchott", 18.098200, -15.948500}
	case "NKG":
		return Location{"Nanjing", "Asia/Shanghai", 31.742001, 118.862000}
	case "NKM":
		return Location{"Nagoya", "Asia/Tokyo", 35.255001, 136.923996}
	case "NLA":
		return Location{"Ndola", "Africa/Lusaka", -12.998100, 28.664900}
	case "NLD":
		return Location{"Nuevo Laredo", "America/Matamoros", 27.443899, -99.570503}
	case "NLF":
		return Location{"Darnley Island", "Australia/Brisbane", -9.583330, 143.766998}
	case "NLG":
		return Location{"Nelson Lagoon", "America/Anchorage", 56.007500, -161.160004}
	case "NLI":
		return Location{"Nikolayevsk-na-Amure Airport", "Asia/Vladivostok", 53.154999, 140.649994}
	case "NLK":
		return Location{"Burnt Pine", "Pacific/Norfolk", -29.041599, 167.938995}
	case "NLU":
		return Location{"Santa Lucia", "America/Mexico_City", 19.756667, -99.015278}
	case "NMA":
		return Location{"Namangan", "Asia/Tashkent", 40.984600, 71.556702}
	case "NME":
		return Location{"Nightmute", "America/Nome", 60.471001, -164.701004}
	case "NNB":
		return Location{"Santa Ana Island", "Pacific/Guadalcanal", -10.847994, 162.454108}
	case "NNG":
		return Location{"Nanning", "Asia/Shanghai", 22.608299, 108.171997}
	case "NNL":
		return Location{"Nondalton", "America/Anchorage", 59.980202, -154.839005}
	case "NNM":
		return Location{"Naryan Mar", "Europe/Moscow", 67.639999, 53.121899}
	case "NNT":
		return Location{"", "Asia/Bangkok", 18.807899, 100.782997}
	case "NNY":
		return Location{"Nanyang", "Asia/Shanghai", 32.980801, 112.614998}
	case "NOB":
		return Location{"Nicoya", "America/Costa_Rica", 9.976490, -85.653000}
	case "NOC":
		return Location{"Charleston", "Europe/Dublin", 53.910301, -8.818490}
	case "NOJ":
		return Location{"Noyabrsk", "Asia/Yekaterinburg", 63.183300, 75.269997}
	case "NOP":
		return Location{"Sinop", "Europe/Istanbul", 42.015800, 35.066399}
	case "NOS":
		return Location{"Nosy Be", "Indian/Antananarivo", -13.312100, 48.314800}
	case "NOU":
		return Location{"Noumea", "Pacific/Noumea", -22.014601, 166.212997}
	case "NOV":
		return Location{"Huambo", "Africa/Luanda", -12.808900, 15.760500}
	case "NOZ":
		return Location{"Novokuznetsk", "Asia/Novokuznetsk", 53.811401, 86.877197}
	case "NPE":
		return Location{"", "Pacific/Auckland", -39.465801, 176.869995}
	case "NPL":
		return Location{"New Plymouth", "Pacific/Auckland", -39.008598, 174.179001}
	case "NQN":
		return Location{"Neuquen", "America/Argentina/Salta", -38.949001, -68.155701}
	case "NQU":
		return Location{"Nuqui", "America/Bogota", 5.696400, -77.280600}
	case "NQY":
		return Location{"Newquay", "Europe/London", 50.440601, -4.995410}
	case "NQZ":
		return Location{"Astana", "Asia/Almaty", 51.022202, 71.466904}
	case "NRA":
		return Location{"Narrandera", "Australia/Sydney", -34.702202, 146.511993}
	case "NRK":
		return Location{"Norrkoping", "Europe/Stockholm", 58.586300, 16.250601}
	case "NRN":
		return Location{"Weeze", "Europe/Amsterdam", 51.602402, 6.142170}
	case "NRR":
		return Location{"Ceiba", "America/Puerto_Rico", 18.245300, -65.643402}
	case "NRT":
		return Location{"Tokyo", "Asia/Tokyo", 35.764702, 140.386002}
	case "NSH":
		return Location{"", "Asia/Tehran", 36.663300, 51.464699}
	case "NSI":
		return Location{"Yaounde", "Africa/Douala", 3.722560, 11.553300}
	case "NSK":
		return Location{"Norilsk", "Asia/Krasnoyarsk", 69.311096, 87.332199}
	case "NSN":
		return Location{"Nelson", "Pacific/Auckland", -41.298302, 173.220993}
	case "NST":
		return Location{"Nakhon Si Thammarat", "Asia/Bangkok", 8.539620, 99.944702}
	case "NTE":
		return Location{"Nantes", "Europe/Paris", 47.153198, -1.610730}
	case "NTG":
		return Location{"Nantong", "Asia/Shanghai", 32.070801, 120.975998}
	case "NTL":
		return Location{"Williamtown", "Australia/Sydney", -32.794998, 151.834000}
	case "NTN":
		return Location{"", "Australia/Brisbane", -17.683599, 141.070007}
	case "NTQ":
		return Location{"Wajima", "Asia/Tokyo", 37.293098, 136.962006}
	case "NTX":
		return Location{"Ranai-Natuna Besar Island", "Asia/Jakarta", 3.908710, 108.388000}
	case "NUE":
		return Location{"Nuremberg", "Europe/Berlin", 49.498699, 11.066900}
	case "NUI":
		return Location{"Nuiqsut", "America/Anchorage", 70.209999, -151.005997}
	case "NUK":
		return Location{"Nukutavake", "Pacific/Tahiti", -19.285000, -138.772003}
	case "NUL":
		return Location{"Nulato", "America/Anchorage", 64.729301, -158.074005}
	case "NUS":
		return Location{"Norsup", "Pacific/Efate", -16.079700, 167.401001}
	case "NUX":
		return Location{"Novy Urengoy", "Asia/Yekaterinburg", 66.069397, 76.520302}
	case "NVA":
		return Location{"Neiva", "America/Bogota", 2.950150, -75.294000}
	case "NVI":
		return Location{"Navoi", "Asia/Samarkand", 40.117199, 65.170799}
	case "NVT":
		return Location{"Navegantes", "America/Sao_Paulo", -26.879999, -48.651402}
	case "NWI":
		return Location{"Norwich", "Europe/London", 52.675800, 1.282780}
	case "NYA":
		return Location{"Nyagan", "Asia/Yekaterinburg", 62.110001, 65.614998}
	case "NYI":
		return Location{"Sunyani", "Africa/Accra", 7.361830, -2.328760}
	case "NYK":
		return Location{"Nanyuki", "Africa/Nairobi", -0.062399, 37.041008}
	case "NYM":
		return Location{"Nadym", "Asia/Yekaterinburg", 65.480904, 72.698898}
	case "NYO":
		return Location{"Stockholm / Nykoping", "Europe/Stockholm", 58.788601, 16.912201}
	case "NYR":
		return Location{"Nyurba", "Asia/Yakutsk", 63.294998, 118.336998}
	case "NYT":
		return Location{"Pyinmana", "Asia/Yangon", 19.623501, 96.200996}
	case "NYU":
		return Location{"Nyaung U", "Asia/Yangon", 21.178801, 94.930199}
	case "NYW":
		return Location{"Monywar", "Asia/Yangon", 22.233000, 95.116997}
	case "OAG":
		return Location{"Orange", "Australia/Sydney", -33.381699, 149.132996}
	case "OAJ":
		return Location{"Jacksonville", "America/New_York", 34.829201, -77.612099}
	case "OAK":
		return Location{"Oakland", "America/Los_Angeles", 37.721298, -122.221001}
	case "OAL":
		return Location{"Cacoal", "America/Porto_Velho", -11.493611, -61.175556}
	case "OAX":
		return Location{"Oaxaca", "America/Mexico_City", 16.999901, -96.726601}
	case "OBO":
		return Location{"Obihiro", "Asia/Tokyo", 42.733299, 143.216995}
	case "OBU":
		return Location{"Kobuk", "America/Anchorage", 66.912300, -156.897003}
	case "OCC":
		return Location{"Coca", "America/Guayaquil", -0.462886, -76.986801}
	case "OCJ":
		return Location{"Ocho Rios", "America/Jamaica", 18.404200, -76.969002}
	case "ODB":
		return Location{"Cordoba", "Europe/Madrid", 37.841999, -4.848880}
	case "ODN":
		return Location{"Long Seridan", "Asia/Kuching", 3.967000, 115.050003}
	case "ODO":
		return Location{"Bodaybo", "Asia/Irkutsk", 57.866100, 114.242996}
	case "ODY":
		return Location{"Oudomsay", "Asia/Vientiane", 20.682699, 101.994003}
	case "OER":
		return Location{"Ornskoldsvik", "Europe/Stockholm", 63.408298, 18.990000}
	case "OGD":
		return Location{"Ogden", "America/Denver", 41.195900, -112.012001}
	case "OGG":
		return Location{"Kahului", "Pacific/Honolulu", 20.898600, -156.429993}
	case "OGL":
		return Location{"Ogle", "America/Guyana", 6.806280, -58.105900}
	case "OGS":
		return Location{"Ogdensburg", "America/New_York", 44.681900, -75.465500}
	case "OGU":
		return Location{"Ordu", "Europe/Istanbul", 40.966667, 38.080000}
	case "OGX":
		return Location{"Ouargla", "Africa/Algiers", 31.917200, 5.412780}
	case "OGZ":
		return Location{"Beslan", "Europe/Moscow", 43.205101, 44.606602}
	case "OHD":
		return Location{"Ohrid", "Europe/Skopje", 41.180000, 20.742300}
	case "OHE":
		return Location{"Mohe", "Asia/Shanghai", 52.912778, 122.430000}
	case "OHH":
		return Location{"Okha", "Asia/Sakhalin", 53.520000, 142.889999}
	case "OHO":
		return Location{"Okhotsk", "Asia/Vladivostok", 59.410065, 143.056503}
	case "OHS":
		return Location{"Sohar", "Asia/Muscat", 24.464167, 56.628333}
	case "OIR":
		return Location{"", "Asia/Tokyo", 42.071701, 139.432999}
	case "OIT":
		return Location{"Oita", "Asia/Tokyo", 33.479401, 131.737000}
	case "OKA":
		return Location{"Naha", "Asia/Tokyo", 26.195801, 127.646004}
	case "OKC":
		return Location{"Oklahoma City", "America/Chicago", 35.393101, -97.600700}
	case "OKD":
		return Location{"Sapporo", "Asia/Tokyo", 43.116100, 141.380005}
	case "OKE":
		return Location{"", "Asia/Tokyo", 27.425501, 128.701004}
	case "OKI":
		return Location{"Okinoshima", "Asia/Tokyo", 36.181099, 133.324997}
	case "OKJ":
		return Location{"Okayama City", "Asia/Tokyo", 34.756901, 133.854996}
	case "OKR":
		return Location{"Yorke Island", "Australia/Brisbane", -9.757030, 143.410995}
	case "OLA":
		return Location{"Orland", "Europe/Oslo", 63.698898, 9.604000}
	case "OLB":
		return Location{"Olbia", "Europe/Rome", 40.898701, 9.517630}
	case "OLF":
		return Location{"Wolf Point", "America/Denver", 48.094501, -105.574997}
	case "OLP":
		return Location{"Olympic Dam", "Australia/Adelaide", -30.485001, 136.876999}
	case "OLZ":
		return Location{"Olyokminsk", "Asia/Yakutsk", 60.397499, 120.471001}
	case "OMA":
		return Location{"Omaha", "America/Chicago", 41.303200, -95.894096}
	case "OMD":
		return Location{"Oranjemund", "Africa/Johannesburg", -28.584700, 16.446699}
	case "OME":
		return Location{"Nome", "America/Nome", 64.512199, -165.445007}
	case "OMH":
		return Location{"Urmia", "Asia/Tehran", 37.668098, 45.068699}
	case "OMO":
		return Location{"Mostar", "Europe/Sarajevo", 43.282902, 17.845900}
	case "OMR":
		return Location{"Oradea", "Europe/Bucharest", 47.025299, 21.902500}
	case "OMS":
		return Location{"Omsk", "Asia/Omsk", 54.966999, 73.310501}
	case "OND":
		return Location{"Ondangwa", "Africa/Windhoek", -17.878201, 15.952600}
	case "ONG":
		return Location{"", "Australia/Brisbane", -16.662500, 139.177994}
	case "ONJ":
		return Location{"Odate", "Asia/Tokyo", 40.191898, 140.371002}
	case "ONK":
		return Location{"Olenyok", "Asia/Yakutsk", 68.514999, 112.480003}
	case "ONQ":
		return Location{"Zonguldak", "Europe/Istanbul", 41.506401, 32.088600}
	case "ONS":
		return Location{"", "Australia/Perth", -21.668301, 115.112999}
	case "ONT":
		return Location{"Ontario", "America/Los_Angeles", 34.056000, -117.600998}
	case "OOK":
		return Location{"Toksook Bay", "America/Nome", 60.541401, -165.087006}
	case "OOL":
		return Location{"Gold Coast", "Australia/Brisbane", -28.164400, 153.505005}
	case "OPF":
		return Location{"Miami", "America/New_York", 25.907000, -80.278397}
	case "OPO":
		return Location{"Porto", "Europe/Lisbon", 41.248100, -8.681390}
	case "OPS":
		return Location{"Sinop", "America/Cuiaba", -11.885000, -55.586109}
	case "ORB":
		return Location{"Orebro", "Europe/Stockholm", 59.223701, 15.038000}
	case "ORD":
		return Location{"Chicago", "America/Chicago", 41.978600, -87.904800}
	case "ORF":
		return Location{"Norfolk", "America/New_York", 36.894600, -76.201202}
	case "ORH":
		return Location{"Worcester", "America/New_York", 42.267300, -71.875702}
	case "ORK":
		return Location{"Cork", "Europe/Dublin", 51.841301, -8.491110}
	case "ORN":
		return Location{"Oran", "Africa/Algiers", 35.623901, -0.621183}
	case "ORT":
		return Location{"Northway", "America/Anchorage", 62.961300, -141.929001}
	case "ORU":
		return Location{"Oruro", "America/La_Paz", -17.962601, -67.076202}
	case "ORV":
		return Location{"Noorvik", "America/Anchorage", 66.817902, -161.018997}
	case "ORX":
		return Location{"Oriximina", "America/Santarem", -1.714080, -55.836201}
	case "ORY":
		return Location{"Paris", "Europe/Paris", 48.725300, 2.359440}
	case "OSD":
		return Location{"Ostersund", "Europe/Stockholm", 63.194401, 14.500300}
	case "OSI":
		return Location{"Osijek", "Europe/Zagreb", 45.462700, 18.810200}
	case "OSL":
		return Location{"Oslo", "Europe/Oslo", 60.193901, 11.100400}
	case "OSR":
		return Location{"Ostrava", "Europe/Prague", 49.696301, 18.111099}
	case "OSS":
		return Location{"Osh", "Asia/Bishkek", 40.608997, 72.793214}
	case "OST":
		return Location{"Ostend", "Europe/Brussels", 51.198898, 2.862220}
	case "OSY":
		return Location{"Namsos", "Europe/Oslo", 64.472198, 11.578600}
	case "OTH":
		return Location{"North Bend", "America/Los_Angeles", 43.417099, -124.246002}
	case "OTI":
		return Location{"Gotalalamo-Morotai Island", "Asia/Jayapura", 2.045990, 128.324997}
	case "OTP":
		return Location{"Bucharest", "Europe/Bucharest", 44.572201, 26.102200}
	case "OTZ":
		return Location{"Kotzebue", "America/Nome", 66.884697, -162.598999}
	case "OUA":
		return Location{"Ouagadougou", "Africa/Ouagadougou", 12.353200, -1.512420}
	case "OUD":
		return Location{"Oujda", "Africa/Casablanca", 34.787201, -1.923990}
	case "OUI":
		return Location{"", "Asia/Bangkok", 20.257299, 100.436996}
	case "OUL":
		return Location{"Oulu / Oulunsalo", "Europe/Helsinki", 64.930099, 25.354601}
	case "OUZ":
		return Location{"Zouerate", "Africa/Nouakchott", 22.756399, -12.483600}
	case "OVB":
		return Location{"Novosibirsk", "Asia/Novosibirsk", 55.012600, 82.650703}
	case "OVD":
		return Location{"Ranon", "Europe/Madrid", 43.563599, -6.034620}
	case "OVS":
		return Location{"Sovetskiy", "Asia/Yekaterinburg", 61.326622, 63.601913}
	case "OWB":
		return Location{"Owensboro", "America/Chicago", 37.740101, -87.166801}
	case "OXB":
		return Location{"Bissau", "Africa/Bissau", 11.894800, -15.653700}
	case "OZC":
		return Location{"Ozamiz City", "Asia/Manila", 8.178510, 123.842003}
	case "OZG":
		return Location{"Zagora", "Africa/Casablanca", 30.320299, -5.866670}
	case "OZZ":
		return Location{"Ouarzazate", "Africa/Casablanca", 30.939100, -6.909430}
	case "PAB":
		return Location{"", "Asia/Kolkata", 21.988400, 82.111000}
	case "PAC":
		return Location{"Albrook", "America/Panama", 8.973340, -79.555603}
	case "PAD":
		return Location{"Paderborn", "Europe/Berlin", 51.614101, 8.616320}
	case "PAE":
		return Location{"Everett", "America/Los_Angeles", 47.906300, -122.281998}
	case "PAF":
		return Location{"", "Africa/Kampala", 2.202222, 31.554444}
	case "PAG":
		return Location{"Pagadian City", "Asia/Manila", 7.830731, 123.461180}
	case "PAH":
		return Location{"Paducah", "America/Chicago", 37.060799, -88.773804}
	case "PAP":
		return Location{"Port-au-Prince", "America/Port-au-Prince", 18.580000, -72.292503}
	case "PAS":
		return Location{"Paros Island", "Europe/Athens", 37.010300, 25.128099}
	case "PAT":
		return Location{"Patna", "Asia/Kolkata", 25.591299, 85.087997}
	case "PAV":
		return Location{"Paulo Afonso", "America/Bahia", -9.400880, -38.250599}
	case "PBC":
		return Location{"Puebla", "America/Mexico_City", 19.158100, -98.371399}
	case "PBG":
		return Location{"Plattsburgh", "America/New_York", 44.650902, -73.468102}
	case "PBH":
		return Location{"Paro", "Asia/Thimphu", 27.403200, 89.424599}
	case "PBI":
		return Location{"West Palm Beach", "America/New_York", 26.683201, -80.095596}
	case "PBJ":
		return Location{"Paama Island", "Pacific/Efate", -16.438999, 168.257004}
	case "PBM":
		return Location{"Zandery", "America/Paramaribo", 5.452830, -55.187801}
	case "PBO":
		return Location{"Paraburdoo", "Australia/Perth", -23.171101, 117.745003}
	case "PBR":
		return Location{"Puerto Barrios", "America/Guatemala", 15.730900, -88.583801}
	case "PBU":
		return Location{"Putao", "Asia/Yangon", 27.329901, 97.426300}
	case "PBZ":
		return Location{"Plettenberg Bay", "Africa/Johannesburg", -34.090302, 23.327801}
	case "PCL":
		return Location{"Pucallpa", "America/Lima", -8.377940, -74.574303}
	case "PCN":
		return Location{"Picton", "Pacific/Auckland", -41.346100, 173.955994}
	case "PCP":
		return Location{"", "Africa/Sao_Tome", 1.662940, 7.411740}
	case "PCR":
		return Location{"Puerto Carreno", "America/Bogota", 6.184720, -67.493200}
	case "PDA":
		return Location{"Puerto Inirida", "America/Bogota", 3.853530, -67.906200}
	case "PDG":
		return Location{"Ketaping/Padang-Sumatra Island", "Asia/Jakarta", -0.786917, 100.280998}
	case "PDL":
		return Location{"Ponta Delgada", "Atlantic/Azores", 37.741199, -25.697901}
	case "PDP":
		return Location{"Punta del Este", "America/Montevideo", -34.855099, -55.094299}
	case "PDS":
		return Location{"", "America/Matamoros", 28.627399, -100.535004}
	case "PDT":
		return Location{"Pendleton", "America/Los_Angeles", 45.695099, -118.841003}
	case "PDV":
		return Location{"Plovdiv", "Europe/Sofia", 42.067799, 24.850800}
	case "PDX":
		return Location{"Portland", "America/Los_Angeles", 45.588699, -122.598000}
	case "PED":
		return Location{"Pardubice", "Europe/Prague", 50.013401, 15.738600}
	case "PEE":
		return Location{"Perm", "Asia/Yekaterinburg", 57.914501, 56.021198}
	case "PEG":
		return Location{"Perugia", "Europe/Rome", 43.095901, 12.513200}
	case "PEI":
		return Location{"Pereira", "America/Bogota", 4.812670, -75.739500}
	case "PEK":
		return Location{"Beijing", "Asia/Shanghai", 40.080101, 116.584999}
	case "PEM":
		return Location{"Puerto Maldonado", "America/Lima", -12.613600, -69.228600}
	case "PEN":
		return Location{"Penang", "Asia/Kuala_Lumpur", 5.297140, 100.277000}
	case "PER":
		return Location{"Perth", "Australia/Perth", -31.940300, 115.967003}
	case "PES":
		return Location{"Petrozavodsk", "Europe/Moscow", 61.885201, 34.154701}
	case "PET":
		return Location{"Pelotas", "America/Sao_Paulo", -31.718399, -52.327702}
	case "PEU":
		return Location{"Puerto Lempira", "America/Tegucigalpa", 15.262200, -83.781197}
	case "PEW":
		return Location{"Peshawar", "Asia/Karachi", 33.993900, 71.514603}
	case "PEZ":
		return Location{"Penza", "Europe/Moscow", 53.110600, 45.021099}
	case "PFB":
		return Location{"Passo Fundo", "America/Sao_Paulo", -28.243999, -52.326599}
	case "PFO":
		return Location{"Paphos", "Asia/Nicosia", 34.717999, 32.485699}
	case "PGA":
		return Location{"Page", "America/Phoenix", 36.926102, -111.447998}
	case "PGD":
		return Location{"Punta Gorda", "America/New_York", 26.920200, -81.990501}
	case "PGF":
		return Location{"Perpignan/Rivesaltes", "Europe/Paris", 42.740398, 2.870670}
	case "PGH":
		return Location{"Pantnagar", "Asia/Kolkata", 29.033400, 79.473701}
	case "PGK":
		return Location{"Pangkal Pinang-Palaubangka Island", "Asia/Jakarta", -2.162200, 106.139000}
	case "PGV":
		return Location{"Greenville", "America/New_York", 35.635201, -77.385300}
	case "PGZ":
		return Location{"Ponta Grossa", "America/Sao_Paulo", -25.184700, -50.144100}
	case "PHB":
		return Location{"Parnaiba", "America/Fortaleza", -2.893750, -41.731998}
	case "PHC":
		return Location{"Port Harcourt", "Africa/Lagos", 5.015490, 6.949590}
	case "PHE":
		return Location{"Port Hedland", "Australia/Perth", -20.377800, 118.625999}
	case "PHF":
		return Location{"Newport News", "America/New_York", 37.131901, -76.492996}
	case "PHL":
		return Location{"Philadelphia", "America/New_York", 39.871899, -75.241096}
	case "PHO":
		return Location{"Point Hope", "America/Nome", 68.348801, -166.798996}
	case "PHS":
		return Location{"", "Asia/Bangkok", 16.782900, 100.278999}
	case "PHX":
		return Location{"Phoenix", "America/Phoenix", 33.434299, -112.012001}
	case "PIA":
		return Location{"Peoria", "America/Chicago", 40.664200, -89.693298}
	case "PIB":
		return Location{"Hattiesburg/Laurel", "America/Chicago", 31.467100, -89.337097}
	case "PIE":
		return Location{"St Petersburg-Clearwater", "America/New_York", 27.910200, -82.687401}
	case "PIH":
		return Location{"Pocatello", "America/Boise", 42.909801, -112.596001}
	case "PIK":
		return Location{"Glasgow", "Europe/London", 55.509399, -4.586670}
	case "PIN":
		return Location{"Parintins", "America/Manaus", -2.673020, -56.777199}
	case "PIP":
		return Location{"Pilot Point", "America/Anchorage", 57.580399, -157.572006}
	case "PIR":
		return Location{"Pierre", "America/Chicago", 44.382702, -100.286003}
	case "PIS":
		return Location{"Poitiers/Biard", "Europe/Paris", 46.587700, 0.306666}
	case "PIT":
		return Location{"Pittsburgh", "America/New_York", 40.491501, -80.232903}
	case "PIU":
		return Location{"Piura", "America/Lima", -5.205750, -80.616402}
	case "PIX":
		return Location{"Pico Island", "Atlantic/Azores", 38.554298, -28.441299}
	case "PIZ":
		return Location{"Point Lay", "America/Nome", 69.732903, -163.005005}
	case "PJA":
		return Location{"", "Europe/Stockholm", 67.245598, 23.068899}
	case "PJM":
		return Location{"Puerto Jimenez", "America/Costa_Rica", 8.533330, -83.300003}
	case "PKA":
		return Location{"Napaskiak", "America/Anchorage", 60.702900, -161.778000}
	case "PKB":
		return Location{"Parkersburg", "America/New_York", 39.345100, -81.439201}
	case "PKC":
		return Location{"Petropavlovsk-Kamchatsky", "Asia/Kamchatka", 53.167900, 158.453995}
	case "PKE":
		return Location{"Parkes", "Australia/Sydney", -33.131401, 148.238998}
	case "PKN":
		return Location{"Pangkalanbun-Borneo Island", "Asia/Pontianak", -2.705200, 111.672997}
	case "PKP":
		return Location{"", "Pacific/Tahiti", -14.809500, -138.813004}
	case "PKR":
		return Location{"Pokhara", "Asia/Kathmandu", 28.200899, 83.982101}
	case "PKU":
		return Location{"Pekanbaru-Sumatra Island", "Asia/Jakarta", 0.460786, 101.445000}
	case "PKV":
		return Location{"Pskov", "Europe/Moscow", 57.783901, 28.395599}
	case "PKX":
		return Location{"Beijing", "Asia/Shanghai", 39.509167, 116.410556}
	case "PKY":
		return Location{"Palangkaraya-Kalimantan Tengah", "Asia/Pontianak", -2.225130, 113.943001}
	case "PKZ":
		return Location{"Pakse", "Asia/Vientiane", 15.132100, 105.780998}
	case "PLM":
		return Location{"Palembang-Sumatra Island", "Asia/Jakarta", -2.898250, 104.699997}
	case "PLN":
		return Location{"Pellston", "America/Detroit", 45.570900, -84.796700}
	case "PLO":
		return Location{"Port Lincoln", "Australia/Adelaide", -34.605301, 135.880005}
	case "PLQ":
		return Location{"Palanga", "Europe/Vilnius", 55.973202, 21.093901}
	case "PLS":
		return Location{"Providenciales Island", "America/Grand_Turk", 21.773600, -72.265900}
	case "PLW":
		return Location{"Palu-Celebes Island", "Asia/Makassar", -0.918542, 119.910004}
	case "PLX":
		return Location{"Semey", "Asia/Almaty", 50.351389, 80.234444}
	case "PLZ":
		return Location{"Port Elizabeth", "Africa/Johannesburg", -33.984901, 25.617300}
	case "PMA":
		return Location{"Chake", "Africa/Dar_es_Salaam", -5.257260, 39.811401}
	case "PMC":
		return Location{"Puerto Montt", "America/Santiago", -41.438900, -73.094002}
	case "PMF":
		return Location{"Parma", "Europe/Rome", 44.824501, 10.296400}
	case "PMG":
		return Location{"Ponta Pora", "America/Asuncion", -22.549601, -55.702599}
	case "PMI":
		return Location{"Palma De Mallorca", "Europe/Madrid", 39.551701, 2.738810}
	case "PML":
		return Location{"Cold Bay", "America/Anchorage", 56.006001, -160.561005}
	case "PMO":
		return Location{"Palermo", "Europe/Rome", 38.175999, 13.091000}
	case "PMR":
		return Location{"", "Pacific/Auckland", -40.320599, 175.617004}
	case "PMV":
		return Location{"Isla Margarita", "America/Caracas", 10.912603, -63.966599}
	case "PMW":
		return Location{"Palmas", "America/Araguaina", -10.291500, -48.356998}
	case "PMY":
		return Location{"Puerto Madryn", "America/Argentina/Catamarca", -42.759200, -65.102700}
	case "PNA":
		return Location{"Pamplona", "Europe/Madrid", 42.770000, -1.646330}
	case "PNH":
		return Location{"Phnom Penh", "Asia/Phnom_Penh", 11.546600, 104.844002}
	case "PNI":
		return Location{"Pohnpei Island", "Pacific/Pohnpei", 6.985100, 158.209000}
	case "PNK":
		return Location{"Pontianak-Borneo Island", "Asia/Pontianak", -0.150711, 109.403999}
	case "PNL":
		return Location{"Pantelleria", "Europe/Rome", 36.816502, 11.968900}
	case "PNP":
		return Location{"Popondetta", "Pacific/Port_Moresby", -8.804540, 148.309006}
	case "PNQ":
		return Location{"Pune", "Asia/Kolkata", 18.582100, 73.919701}
	case "PNR":
		return Location{"Pointe Noire", "Africa/Brazzaville", -4.816030, 11.886600}
	case "PNS":
		return Location{"Pensacola", "America/Chicago", 30.473400, -87.186600}
	case "PNT":
		return Location{"Puerto Natales", "America/Punta_Arenas", -51.671501, -72.528397}
	case "PNY":
		return Location{"", "Asia/Kolkata", 11.968700, 79.810097}
	case "PNZ":
		return Location{"Petrolina", "America/Recife", -9.362410, -40.569099}
	case "POA":
		return Location{"Porto Alegre", "America/Sao_Paulo", -29.994400, -51.171398}
	case "POG":
		return Location{"Port Gentil", "Africa/Libreville", -0.711739, 8.754380}
	case "POJ":
		return Location{"Patos De Minas", "America/Sao_Paulo", -18.672800, -46.491199}
	case "POL":
		return Location{"Pemba / Porto Amelia", "Africa/Maputo", -12.991762, 40.524014}
	case "POM":
		return Location{"Port Moresby", "Pacific/Port_Moresby", -9.443380, 147.220001}
	case "POP":
		return Location{"Puerto Plata", "America/Santo_Domingo", 19.757900, -70.570000}
	case "POR":
		return Location{"Pori", "Europe/Helsinki", 61.461700, 21.799999}
	case "POS":
		return Location{"Port of Spain", "America/Port_of_Spain", 10.595400, -61.337200}
	case "POZ":
		return Location{"Poznan", "Europe/Warsaw", 52.421001, 16.826300}
	case "PPB":
		return Location{"Presidente Prudente", "America/Sao_Paulo", -22.175100, -51.424599}
	case "PPG":
		return Location{"Pago Pago", "Pacific/Pago_Pago", -14.331000, -170.710007}
	case "PPK":
		return Location{"Petropavlosk", "Asia/Almaty", 54.774700, 69.183899}
	case "PPN":
		return Location{"Popayan", "America/Bogota", 2.454400, -76.609300}
	case "PPP":
		return Location{"Proserpine", "Australia/Brisbane", -20.495001, 148.552002}
	case "PPQ":
		return Location{"", "Pacific/Auckland", -40.904701, 174.988998}
	case "PPS":
		return Location{"Puerto Princesa City", "Asia/Manila", 9.742120, 118.759003}
	case "PPT":
		return Location{"Papeete", "Pacific/Tahiti", -17.553699, -149.606995}
	case "PQC":
		return Location{"Duong Dong", "Asia/Ho_Chi_Minh", 10.227000, 103.967003}
	case "PQI":
		return Location{"Presque Isle", "America/New_York", 46.688999, -68.044800}
	case "PQQ":
		return Location{"Port Macquarie", "Australia/Sydney", -31.435801, 152.863007}
	case "PRA":
		return Location{"Parana", "America/Argentina/Cordoba", -31.794800, -60.480400}
	case "PRC":
		return Location{"Prescott", "America/Phoenix", 34.654499, -112.419998}
	case "PRG":
		return Location{"Prague", "Europe/Prague", 50.100800, 14.260000}
	case "PRI":
		return Location{"Praslin Island", "Indian/Mahe", -4.319290, 55.691399}
	case "PRN":
		return Location{"Prishtina", "Europe/Belgrade", 42.572800, 21.035801}
	case "PRS":
		return Location{"Parasi", "Pacific/Guadalcanal", -9.641670, 161.425003}
	case "PSA":
		return Location{"Pisa", "Europe/Rome", 43.683899, 10.392700}
	case "PSC":
		return Location{"Pasco", "America/Los_Angeles", 46.264702, -119.119003}
	case "PSE":
		return Location{"Ponce", "America/Puerto_Rico", 18.008301, -66.563004}
	case "PSG":
		return Location{"Petersburg", "America/Sitka", 56.801701, -132.945007}
	case "PSM":
		return Location{"Portsmouth", "America/New_York", 43.077900, -70.823303}
	case "PSO":
		return Location{"Pasto", "America/Bogota", 1.396250, -77.291500}
	case "PSP":
		return Location{"Palm Springs", "America/Los_Angeles", 33.829700, -116.507004}
	case "PSR":
		return Location{"Pescara", "Europe/Rome", 42.431702, 14.181100}
	case "PSS":
		return Location{"Posadas", "America/Argentina/Cordoba", -27.385800, -55.970700}
	case "PSU":
		return Location{"Putussibau-Borneo Island", "Asia/Pontianak", 0.835578, 112.936996}
	case "PTA":
		return Location{"Port Alsworth", "America/Anchorage", 60.204300, -154.319000}
	case "PTG":
		return Location{"Potgietersrus", "Africa/Johannesburg", -23.845301, 29.458599}
	case "PTH":
		return Location{"Port Heiden", "America/Anchorage", 56.959099, -158.632996}
	case "PTO":
		return Location{"Pato Branco", "America/Sao_Paulo", -26.217800, -52.694302}
	case "PTP":
		return Location{"Pointe-a-Pitre Le Raizet", "America/Guadeloupe", 16.265301, -61.531799}
	case "PTQ":
		return Location{"Porto De Moz", "America/Belem", -1.741450, -52.236099}
	case "PTU":
		return Location{"Platinum", "America/Anchorage", 59.011398, -161.820007}
	case "PTX":
		return Location{"Pitalito", "America/Bogota", 1.857770, -76.085700}
	case "PTY":
		return Location{"Tocumen", "America/Panama", 9.071360, -79.383499}
	case "PUB":
		return Location{"Pueblo", "America/Denver", 38.289101, -104.497002}
	case "PUF":
		return Location{"Pau/Pyrenees (Uzein)", "Europe/Paris", 43.380001, -0.418611}
	case "PUJ":
		return Location{"Punta Cana", "America/Santo_Domingo", 18.567400, -68.363403}
	case "PUK":
		return Location{"Pukarua", "Pacific/Tahiti", -18.295601, -137.016998}
	case "PUQ":
		return Location{"Punta Arenas", "America/Punta_Arenas", -53.002602, -70.854599}
	case "PUS":
		return Location{"Busan", "Asia/Seoul", 35.179501, 128.938004}
	case "PUU":
		return Location{"Puerto Asis", "America/Bogota", 0.505228, -76.500800}
	case "PUW":
		return Location{"Pullman/Moscow", "America/Los_Angeles", 46.743900, -117.110001}
	case "PUY":
		return Location{"Pula", "Europe/Zagreb", 44.893501, 13.922200}
	case "PVA":
		return Location{"Providencia", "America/Bogota", 13.356900, -81.358300}
	case "PVC":
		return Location{"Provincetown", "America/New_York", 42.071899, -70.221397}
	case "PVD":
		return Location{"Providence", "America/New_York", 41.732601, -71.420403}
	case "PVG":
		return Location{"Shanghai", "Asia/Shanghai", 31.143400, 121.805000}
	case "PVH":
		return Location{"Porto Velho", "America/Porto_Velho", -8.709290, -63.902302}
	case "PVK":
		return Location{"Preveza/Lefkada", "Europe/Athens", 38.925499, 20.765301}
	case "PVR":
		return Location{"Puerto Vallarta", "America/Bahia_Banderas", 20.680099, -105.253998}
	case "PVU":
		return Location{"Provo", "America/Denver", 40.219200, -111.723000}
	case "PWE":
		return Location{"Pevek", "Asia/Anadyr", 69.783302, 170.597000}
	case "PWM":
		return Location{"Portland", "America/New_York", 43.646198, -70.309303}
	case "PWQ":
		return Location{"Pavlodar", "Asia/Almaty", 52.195000, 77.073898}
	case "PXM":
		return Location{"Puerto Escondido", "America/Mexico_City", 15.876900, -97.089104}
	case "PXO":
		return Location{"Vila Baleira", "Atlantic/Madeira", 33.073399, -16.350000}
	case "PXU":
		return Location{"Pleiku", "Asia/Ho_Chi_Minh", 14.004500, 108.016998}
	case "PYB":
		return Location{"Jeypore", "Asia/Kolkata", 18.879999, 82.552002}
	case "PYH":
		return Location{"", "America/Bogota", 5.619990, -67.606102}
	case "PYJ":
		return Location{"Yakutia", "Asia/Yakutsk", 66.400398, 112.029999}
	case "PYM":
		return Location{"Plymouth", "America/New_York", 41.909000, -70.728798}
	case "PZB":
		return Location{"Pietermaritzburg", "Africa/Johannesburg", -29.649000, 30.398701}
	case "PZO":
		return Location{"Puerto Ordaz-Ciudad Guayana", "America/Caracas", 8.288530, -62.760399}
	case "PZU":
		return Location{"Port Sudan", "Africa/Khartoum", 19.433599, 37.234100}
	case "QBC":
		return Location{"Bella Coola", "America/Vancouver", 52.387501, -126.596001}
	case "QGP":
		return Location{"Garanhuns", "America/Recife", -8.834280, -36.471600}
	case "QIG":
		return Location{"Iguatu", "America/Fortaleza", -6.346640, -39.293800}
	case "QOW":
		return Location{"Owerri", "Africa/Lagos", 5.427060, 7.206030}
	case "QPA":
		return Location{"Padova", "Europe/Rome", 45.395802, 11.847900}
	case "QRO":
		return Location{"Queretaro", "America/Mexico_City", 20.617300, -100.185997}
	case "QSF":
		return Location{"Setif", "Africa/Algiers", 36.178101, 5.324490}
	case "QXB":
		return Location{"Lyon", "Europe/Paris", 43.505600, 5.367780}
	case "RAB":
		return Location{"Tokua", "Pacific/Port_Moresby", -4.340460, 152.380005}
	case "RAE":
		return Location{"Arar", "Asia/Riyadh", 30.906601, 41.138199}
	case "RAH":
		return Location{"Rafha", "Asia/Riyadh", 29.626400, 43.490601}
	case "RAI":
		return Location{"Praia", "Atlantic/Cape_Verde", 14.924500, -23.493500}
	case "RAK":
		return Location{"Marrakech", "Africa/Casablanca", 31.606899, -8.036300}
	case "RAO":
		return Location{"Ribeirao Preto", "America/Sao_Paulo", -21.136389, -47.776669}
	case "RAP":
		return Location{"Rapid City", "America/Denver", 44.045300, -103.056999}
	case "RAR":
		return Location{"Avarua", "Pacific/Rarotonga", -21.202700, -159.806000}
	case "RAS":
		return Location{"Rasht", "Asia/Tehran", 37.323333, 49.617778}
	case "RBA":
		return Location{"Rabat", "Africa/Casablanca", 34.051498, -6.751520}
	case "RBB":
		return Location{"Borba", "America/Manaus", -4.406340, -59.602402}
	case "RBR":
		return Location{"Rio Branco", "America/Rio_Branco", -9.868889, -67.898056}
	case "RBV":
		return Location{"Ramata", "Pacific/Guadalcanal", -8.168060, 157.643005}
	case "RBY":
		return Location{"Ruby", "America/Anchorage", 64.727203, -155.470001}
	case "RCB":
		return Location{"Richards Bay", "Africa/Johannesburg", -28.740999, 32.092098}
	case "RCH":
		return Location{"Riohacha", "America/Bogota", 11.526200, -72.926000}
	case "RCM":
		return Location{"", "Australia/Brisbane", -20.701900, 143.115005}
	case "RCQ":
		return Location{"Reconquista", "America/Argentina/Cordoba", -29.210300, -59.680000}
	case "RCU":
		return Location{"Rio Cuarto", "America/Argentina/Cordoba", -33.085098, -64.261299}
	case "RDD":
		return Location{"Redding", "America/Los_Angeles", 40.508999, -122.292999}
	case "RDM":
		return Location{"Redmond", "America/Los_Angeles", 44.254101, -121.150002}
	case "RDO":
		return Location{"Radom", "Europe/Warsaw", 51.389198, 21.213301}
	case "RDU":
		return Location{"Raleigh/Durham", "America/New_York", 35.877602, -78.787498}
	case "RDZ":
		return Location{"Rodez/Marcillac", "Europe/Paris", 44.407902, 2.482670}
	case "REA":
		return Location{"", "Pacific/Tahiti", -18.465900, -136.440002}
	case "REC":
		return Location{"Recife", "America/Recife", -8.126490, -34.923599}
	case "REG":
		return Location{"Reggio Calabria", "Europe/Rome", 38.071201, 15.651600}
	case "REL":
		return Location{"Rawson", "America/Argentina/Catamarca", -43.210500, -65.270300}
	case "REN":
		return Location{"Orenburg", "Asia/Yekaterinburg", 51.795799, 55.456699}
	case "RER":
		return Location{"Retalhuleu", "America/Guatemala", 14.521000, -91.697304}
	case "RES":
		return Location{"Resistencia", "America/Argentina/Cordoba", -27.450000, -59.056100}
	case "RET":
		return Location{"", "Europe/Oslo", 67.527802, 12.103300}
	case "REU":
		return Location{"Reus", "Europe/Madrid", 41.147400, 1.167170}
	case "REX":
		return Location{"Reynosa", "America/Matamoros", 26.008900, -98.228500}
	case "RFD":
		return Location{"Chicago/Rockford", "America/Chicago", 42.195400, -89.097198}
	case "RFP":
		return Location{"Uturoa", "Pacific/Tahiti", -16.722900, -151.466003}
	case "RGA":
		return Location{"Rio Grande", "America/Argentina/Ushuaia", -53.777700, -67.749400}
	case "RGI":
		return Location{"", "Pacific/Tahiti", -14.954300, -147.660995}
	case "RGK":
		return Location{"Gorno-Altaysk", "Asia/Barnaul", 51.966702, 85.833298}
	case "RGL":
		return Location{"Rio Gallegos", "America/Argentina/Rio_Gallegos", -51.608900, -69.312600}
	case "RGN":
		return Location{"Yangon", "Asia/Yangon", 16.907301, 96.133202}
	case "RGS":
		return Location{"Burgos", "Europe/Madrid", 42.357601, -3.620760}
	case "RHD":
		return Location{"Rio Hondo", "America/Argentina/Cordoba", -27.473700, -64.905502}
	case "RHI":
		return Location{"Rhinelander", "America/Chicago", 45.631199, -89.467499}
	case "RHO":
		return Location{"Rodes Island", "Europe/Athens", 36.405399, 28.086201}
	case "RIA":
		return Location{"Santa Maria", "America/Sao_Paulo", -29.711399, -53.688202}
	case "RIC":
		return Location{"Richmond", "America/New_York", 37.505199, -77.319702}
	case "RIG":
		return Location{"Rio Grande", "America/Sao_Paulo", -32.081699, -52.163299}
	case "RIH":
		return Location{"Rio Hato", "America/Panama", 8.375833, -80.127778}
	case "RIS":
		return Location{"Rishiri", "Asia/Tokyo", 45.242001, 141.186005}
	case "RIW":
		return Location{"Riverton", "America/Denver", 43.064201, -108.459999}
	case "RIX":
		return Location{"Riga", "Europe/Riga", 56.923599, 23.971100}
	case "RJA":
		return Location{"Rajahmundry", "Asia/Kolkata", 17.110399, 81.818199}
	case "RJB":
		return Location{"Rajbiraj", "Asia/Kathmandu", 26.517000, 86.750000}
	case "RJH":
		return Location{"Rajshahi", "Asia/Dhaka", 24.437201, 88.616501}
	case "RJK":
		return Location{"Rijeka", "Europe/Zagreb", 45.216900, 14.570300}
	case "RJL":
		return Location{"Logrono", "Europe/Madrid", 42.460953, -2.322235}
	case "RKA":
		return Location{"", "Pacific/Tahiti", -15.485300, -145.470001}
	case "RKD":
		return Location{"Rockland", "America/New_York", 44.060101, -69.099197}
	case "RKS":
		return Location{"Rock Springs", "America/Denver", 41.594200, -109.065002}
	case "RKT":
		return Location{"Ras Al Khaimah", "Asia/Dubai", 25.613501, 55.938801}
	case "RKV":
		return Location{"Reykjavik", "Atlantic/Reykjavik", 64.129997, -21.940599}
	case "RLG":
		return Location{"Rostock", "Europe/Berlin", 53.918201, 12.278300}
	case "RLO":
		return Location{"Merlo", "America/Argentina/San_Luis", -32.384701, -65.186501}
	case "RMA":
		return Location{"Roma", "Australia/Brisbane", -26.545000, 148.774994}
	case "RMF":
		return Location{"Marsa Alam", "Africa/Cairo", 25.557100, 34.583698}
	case "RMI":
		return Location{"Rimini", "Europe/Rome", 44.020302, 12.611700}
	case "RMQ":
		return Location{"Taichung City", "Asia/Taipei", 24.264700, 120.621002}
	case "RMU":
		return Location{"Corvera", "Europe/Madrid", 37.803000, -1.125000}
	case "RNA":
		return Location{"Arona", "Pacific/Guadalcanal", -9.860544, 161.979547}
	case "RNB":
		return Location{"", "Europe/Stockholm", 56.266701, 15.265000}
	case "RNJ":
		return Location{"", "Asia/Tokyo", 27.044001, 128.401993}
	case "RNL":
		return Location{"Rennell Island", "Pacific/Guadalcanal", -11.533900, 160.063004}
	case "RNN":
		return Location{"Ronne", "Europe/Copenhagen", 55.063301, 14.759600}
	case "RNO":
		return Location{"Reno", "America/Los_Angeles", 39.499100, -119.767998}
	case "RNS":
		return Location{"Rennes/Saint-Jacques", "Europe/Paris", 48.069500, -1.734790}
	case "ROA":
		return Location{"Roanoke", "America/New_York", 37.325500, -79.975403}
	case "ROB":
		return Location{"Monrovia", "Africa/Monrovia", 6.233790, -10.362300}
	case "ROC":
		return Location{"Rochester", "America/New_York", 43.118900, -77.672401}
	case "ROI":
		return Location{"", "Asia/Bangkok", 16.116800, 103.774002}
	case "ROK":
		return Location{"Rockhampton", "Australia/Brisbane", -23.381901, 150.475006}
	case "RON":
		return Location{"Paipa", "America/Bogota", 5.764540, -73.105400}
	case "ROO":
		return Location{"Rondonopolis", "America/Cuiaba", -16.586000, -54.724800}
	case "ROR":
		return Location{"Babelthuap Island", "Pacific/Palau", 7.367650, 134.544006}
	case "ROS":
		return Location{"Rosario", "America/Argentina/Cordoba", -32.903600, -60.785000}
	case "ROT":
		return Location{"Rotorua", "Pacific/Auckland", -38.109200, 176.317001}
	case "ROV":
		return Location{"Rostov-on-Don", "Europe/Moscow", 47.500833, 39.933611}
	case "ROW":
		return Location{"Roswell", "America/Denver", 33.301601, -104.530998}
	case "RPR":
		return Location{"Raipur", "Asia/Kolkata", 21.180401, 81.738800}
	case "RRG":
		return Location{"Port Mathurin", "Indian/Mauritius", -19.757700, 63.361000}
	case "RRK":
		return Location{"", "Asia/Kolkata", 22.256701, 84.814598}
	case "RRR":
		return Location{"", "Pacific/Tahiti", -16.045000, -142.476944}
	case "RRS":
		return Location{"Roros", "Europe/Oslo", 62.578400, 11.342300}
	case "RSA":
		return Location{"Santa Rosa", "America/Argentina/Salta", -36.588299, -64.275703}
	case "RSD":
		return Location{"Rock Sound", "America/Nassau", 24.895079, -76.176882}
	case "RSH":
		return Location{"Russian Mission", "America/Anchorage", 61.778885, -161.319458}
	case "RST":
		return Location{"Rochester", "America/Chicago", 43.908298, -92.500000}
	case "RSU":
		return Location{"Yeosu", "Asia/Seoul", 34.842300, 127.616997}
	case "RSW":
		return Location{"Fort Myers", "America/New_York", 26.536200, -81.755203}
	case "RTA":
		return Location{"Rotuma", "Pacific/Fiji", -12.482500, 177.070999}
	case "RTB":
		return Location{"Roatan Island", "America/Tegucigalpa", 16.316799, -86.523003}
	case "RTG":
		return Location{"Satar Tacik-Flores Island", "Asia/Makassar", -8.597010, 120.476997}
	case "RTM":
		return Location{"Rotterdam", "Europe/Amsterdam", 51.956902, 4.437220}
	case "RUH":
		return Location{"Riyadh", "Asia/Riyadh", 24.957600, 46.698799}
	case "RUN":
		return Location{"St Denis", "Indian/Reunion", -20.887100, 55.510300}
	case "RUR":
		return Location{"", "Pacific/Tahiti", -22.434099, -151.360992}
	case "RUT":
		return Location{"Rutland", "America/New_York", 43.529400, -72.949600}
	case "RVD":
		return Location{"Rio Verde", "America/Sao_Paulo", -17.834723, -50.956112}
	case "RVE":
		return Location{"Saravena", "America/Bogota", 6.951389, -71.856944}
	case "RVK":
		return Location{"Rorvik", "Europe/Oslo", 64.838303, 11.146100}
	case "RVN":
		return Location{"Rovaniemi", "Europe/Helsinki", 66.564796, 25.830400}
	case "RVV":
		return Location{"", "Pacific/Tahiti", -23.885201, -147.662003}
	case "RXS":
		return Location{"Roxas City", "Asia/Manila", 11.597700, 122.751999}
	case "RZE":
		return Location{"Rzeszow", "Europe/Warsaw", 50.110001, 22.018999}
	case "RZR":
		return Location{"", "Asia/Tehran", 36.909901, 50.679600}
	case "SAB":
		return Location{"Saba", "America/Kralendijk", 17.645000, -63.220001}
	case "SAF":
		return Location{"Santa Fe", "America/Denver", 35.617100, -106.088997}
	case "SAG":
		return Location{"Kakadi", "Asia/Kolkata", 19.688611, 74.378889}
	case "SAL":
		return Location{"Santa Clara", "America/El_Salvador", 13.440900, -89.055702}
	case "SAN":
		return Location{"San Diego", "America/Los_Angeles", 32.733601, -117.190002}
	case "SAP":
		return Location{"La Mesa", "America/Tegucigalpa", 15.452600, -87.923599}
	case "SAQ":
		return Location{"Andros Island", "America/Nassau", 25.053801, -78.049004}
	case "SAT":
		return Location{"San Antonio", "America/Chicago", 29.533701, -98.469803}
	case "SAV":
		return Location{"Savannah", "America/New_York", 32.127602, -81.202103}
	case "SAW":
		return Location{"Istanbul", "Europe/Istanbul", 40.898602, 29.309200}
	case "SBA":
		return Location{"Santa Barbara", "America/Los_Angeles", 34.426201, -119.839996}
	case "SBD":
		return Location{"San Bernardino", "America/Los_Angeles", 34.095402, -117.235001}
	case "SBH":
		return Location{"Gustavia", "America/St_Barthelemy", 17.904400, -62.843601}
	case "SBN":
		return Location{"South Bend", "America/Indiana/Indianapolis", 41.708698, -86.317299}
	case "SBP":
		return Location{"San Luis Obispo", "America/Los_Angeles", 35.236801, -120.641998}
	case "SBR":
		return Location{"Saibai Island", "Australia/Brisbane", -9.378330, 142.625000}
	case "SBW":
		return Location{"Sibu", "Asia/Kuching", 2.261600, 111.985001}
	case "SBY":
		return Location{"Salisbury", "America/New_York", 38.340500, -75.510300}
	case "SBZ":
		return Location{"Sibiu", "Europe/Bucharest", 45.785599, 24.091299}
	case "SCC":
		return Location{"Deadhorse", "America/Anchorage", 70.194702, -148.464996}
	case "SCE":
		return Location{"State College", "America/New_York", 40.849300, -77.848701}
	case "SCF":
		return Location{"Scottsdale", "America/Phoenix", 33.622898, -111.911003}
	case "SCK":
		return Location{"Stockton", "America/Los_Angeles", 37.894199, -121.237999}
	case "SCL":
		return Location{"Santiago", "America/Santiago", -33.393002, -70.785797}
	case "SCM":
		return Location{"Scammon Bay", "America/Nome", 61.845299, -165.570999}
	case "SCN":
		return Location{"Saarbrucken", "Europe/Berlin", 49.214600, 7.109510}
	case "SCO":
		return Location{"Aktau", "Asia/Aqtau", 43.860100, 51.091999}
	case "SCQ":
		return Location{"Santiago de Compostela", "Europe/Madrid", 42.896301, -8.415140}
	case "SCU":
		return Location{"Santiago", "America/Havana", 19.969801, -75.835403}
	case "SCV":
		return Location{"Suceava", "Europe/Bucharest", 47.687500, 26.354099}
	case "SCW":
		return Location{"Syktyvkar", "Europe/Moscow", 61.646999, 50.845100}
	case "SCY":
		return Location{"San Cristobal", "Pacific/Galapagos", -0.910206, -89.617401}
	case "SCZ":
		return Location{"Santa Cruz/Graciosa Bay/Luova", "Pacific/Guadalcanal", -10.720300, 165.794998}
	case "SDD":
		return Location{"Lubango", "Africa/Luanda", -14.924700, 13.575000}
	case "SDE":
		return Location{"Santiago del Estero", "America/Argentina/Cordoba", -27.765556, -64.309998}
	case "SDF":
		return Location{"Louisville", "America/Kentucky/Louisville", 38.174400, -85.736000}
	case "SDG":
		return Location{"", "Asia/Tehran", 35.245899, 47.009201}
	case "SDJ":
		return Location{"Sendai", "Asia/Tokyo", 38.139702, 140.917007}
	case "SDK":
		return Location{"Sandakan", "Asia/Kuching", 5.900900, 118.058998}
	case "SDL":
		return Location{"Sundsvall/ Harnosand", "Europe/Stockholm", 62.528099, 17.443899}
	case "SDN":
		return Location{"Sandane", "Europe/Oslo", 61.830002, 6.105830}
	case "SDP":
		return Location{"Sand Point", "America/Anchorage", 55.314999, -160.522995}
	case "SDQ":
		return Location{"Santo Domingo", "America/Santo_Domingo", 18.429701, -69.668900}
	case "SDR":
		return Location{"Santander", "Europe/Madrid", 43.427101, -3.820010}
	case "SDU":
		return Location{"Rio De Janeiro", "America/Sao_Paulo", -22.910500, -43.163101}
	case "SDY":
		return Location{"Sidney", "America/Denver", 47.706902, -104.193001}
	case "SEA":
		return Location{"Seattle", "America/Los_Angeles", 47.449001, -122.308998}
	case "SEB":
		return Location{"Sabha", "Africa/Tripoli", 26.987000, 14.472500}
	case "SEN":
		return Location{"Southend", "Europe/London", 51.571400, 0.695556}
	case "SEU":
		return Location{"Seronera", "Africa/Dar_es_Salaam", -2.458060, 34.822498}
	case "SEZ":
		return Location{"Mahe Island", "Indian/Mahe", -4.674340, 55.521801}
	case "SFA":
		return Location{"Sfax", "Africa/Tunis", 34.717999, 10.691000}
	case "SFB":
		return Location{"Orlando", "America/New_York", 28.777599, -81.237503}
	case "SFD":
		return Location{"Inglaterra", "America/Caracas", 7.883320, -67.444000}
	case "SFG":
		return Location{"Grand Case", "America/Lower_Princes", 18.099899, -63.047199}
	case "SFJ":
		return Location{"Kangerlussuaq", "America/Nuuk", 67.012222, -50.711603}
	case "SFL":
		return Location{"Sao Filipe", "Atlantic/Cape_Verde", 14.885000, -24.480000}
	case "SFN":
		return Location{"Santa Fe", "America/Argentina/Cordoba", -31.711700, -60.811700}
	case "SFO":
		return Location{"San Francisco", "America/Los_Angeles", 37.618999, -122.375000}
	case "SFT":
		return Location{"Skelleftea", "Europe/Stockholm", 64.624802, 21.076900}
	case "SGC":
		return Location{"Surgut", "Asia/Yekaterinburg", 61.343700, 73.401802}
	case "SGD":
		return Location{"Sonderborg", "Europe/Copenhagen", 54.964401, 9.791730}
	case "SGF":
		return Location{"Springfield", "America/Chicago", 37.245701, -93.388603}
	case "SGN":
		return Location{"Ho Chi Minh City", "Asia/Ho_Chi_Minh", 10.818800, 106.652000}
	case "SGO":
		return Location{"", "Australia/Brisbane", -28.049700, 148.595001}
	case "SGU":
		return Location{"St George", "America/Denver", 37.036389, -113.510306}
	case "SGX":
		return Location{"Songea", "Africa/Dar_es_Salaam", -10.683000, 35.583000}
	case "SGY":
		return Location{"Skagway", "America/Juneau", 59.460098, -135.315994}
	case "SHA":
		return Location{"Shanghai", "Asia/Shanghai", 31.197901, 121.335999}
	case "SHB":
		return Location{"Nakashibetsu", "Asia/Tokyo", 43.577499, 144.960007}
	case "SHC":
		return Location{"Shire", "Africa/Addis_Ababa", 14.079444, 38.270833}
	case "SHD":
		return Location{"Staunton/Waynesboro/Harrisonburg", "America/New_York", 38.263802, -78.896400}
	case "SHE":
		return Location{"Shenyang", "Asia/Shanghai", 41.639801, 123.483002}
	case "SHH":
		return Location{"Shishmaref", "America/Nome", 66.249603, -166.089005}
	case "SHI":
		return Location{"", "Asia/Tokyo", 24.826700, 125.144997}
	case "SHJ":
		return Location{"Sharjah", "Asia/Dubai", 25.328600, 55.517200}
	case "SHL":
		return Location{"Shillong", "Asia/Kolkata", 25.703600, 91.978699}
	case "SHM":
		return Location{"Shirahama", "Asia/Tokyo", 33.662201, 135.363998}
	case "SHO":
		return Location{"", "Asia/Seoul", 38.142601, 128.598999}
	case "SHR":
		return Location{"Sheridan", "America/Denver", 44.769199, -106.980003}
	case "SHS":
		return Location{"Shashi", "Asia/Shanghai", 30.324400, 112.280998}
	case "SHV":
		return Location{"Shreveport", "America/Chicago", 32.446602, -93.825600}
	case "SHW":
		return Location{"", "Asia/Riyadh", 17.466900, 47.121399}
	case "SHX":
		return Location{"Shageluk", "America/Anchorage", 62.692299, -159.569000}
	case "SID":
		return Location{"Espargos", "Atlantic/Cape_Verde", 16.741400, -22.949400}
	case "SIF":
		return Location{"Simara", "Asia/Kathmandu", 27.159500, 84.980103}
	case "SIG":
		return Location{"San Juan", "America/Puerto_Rico", 18.456800, -66.098099}
	case "SIN":
		return Location{"Singapore", "Asia/Singapore", 1.350190, 103.994003}
	case "SIR":
		return Location{"Sion", "Europe/Zurich", 46.219601, 7.326760}
	case "SIS":
		return Location{"Sishen", "Africa/Johannesburg", -27.648600, 22.999300}
	case "SIT":
		return Location{"Sitka", "America/Sitka", 57.047100, -135.362000}
	case "SJC":
		return Location{"San Jose", "America/Los_Angeles", 37.362598, -121.929001}
	case "SJD":
		return Location{"San Jose del Cabo", "America/Mazatlan", 23.151800, -109.721001}
	case "SJE":
		return Location{"San Jose Del Guaviare", "America/Bogota", 2.579690, -72.639400}
	case "SJI":
		return Location{"San Jose", "Asia/Manila", 12.361500, 121.046997}
	case "SJJ":
		return Location{"Sarajevo", "Europe/Sarajevo", 43.824600, 18.331499}
	case "SJK":
		return Location{"Sao Jose Dos Campos", "America/Sao_Paulo", -23.229200, -45.861500}
	case "SJL":
		return Location{"Sao Gabriel Da Cachoeira", "America/Manaus", -0.148350, -66.985500}
	case "SJO":
		return Location{"San Jose", "America/Costa_Rica", 9.993860, -84.208801}
	case "SJP":
		return Location{"Sao Jose Do Rio Preto", "America/Sao_Paulo", -20.816601, -49.406502}
	case "SJT":
		return Location{"San Angelo", "America/Chicago", 31.357700, -100.496002}
	case "SJU":
		return Location{"San Juan", "America/Puerto_Rico", 18.439400, -66.001801}
	case "SJW":
		return Location{"Shijiazhuang", "Asia/Shanghai", 38.280701, 114.696999}
	case "SJZ":
		return Location{"Velas", "Atlantic/Azores", 38.665501, -28.175800}
	case "SKB":
		return Location{"Basseterre", "America/St_Kitts", 17.311199, -62.718700}
	case "SKD":
		return Location{"Samarkand", "Asia/Samarkand", 39.700500, 66.983803}
	case "SKG":
		return Location{"Thessaloniki", "Europe/Athens", 40.519699, 22.970900}
	case "SKH":
		return Location{"Surkhet", "Asia/Kathmandu", 28.586000, 81.636002}
	case "SKK":
		return Location{"Shaktoolik", "America/Anchorage", 64.371101, -161.223999}
	case "SKN":
		return Location{"Hadsel", "Europe/Oslo", 68.578827, 15.033417}
	case "SKO":
		return Location{"Sokoto", "Africa/Lagos", 12.916300, 5.207190}
	case "SKP":
		return Location{"Skopje", "Europe/Skopje", 41.961601, 21.621401}
	case "SKT":
		return Location{"Sialkot", "Asia/Karachi", 32.535557, 74.363892}
	case "SKU":
		return Location{"Skiros Island", "Europe/Athens", 38.967602, 24.487200}
	case "SKX":
		return Location{"Saransk", "Europe/Moscow", 54.125130, 45.212257}
	case "SKZ":
		return Location{"Mirpur Khas", "Asia/Karachi", 27.722000, 68.791702}
	case "SLA":
		return Location{"Salta", "America/Argentina/Salta", -24.856001, -65.486198}
	case "SLC":
		return Location{"Salt Lake City", "America/Denver", 40.788399, -111.977997}
	case "SLE":
		return Location{"Salem", "America/Los_Angeles", 44.909500, -123.002998}
	case "SLH":
		return Location{"Sola", "Pacific/Efate", -13.851700, 167.537003}
	case "SLI":
		return Location{"Solwesi", "Africa/Lusaka", -12.173700, 26.365101}
	case "SLK":
		return Location{"Saranac Lake", "America/New_York", 44.385300, -74.206200}
	case "SLL":
		return Location{"Salalah", "Asia/Muscat", 17.038700, 54.091301}
	case "SLM":
		return Location{"Salamanca", "Europe/Madrid", 40.952099, -5.501990}
	case "SLN":
		return Location{"Salina", "America/Chicago", 38.791000, -97.652199}
	case "SLP":
		return Location{"San Luis Potosi", "America/Mexico_City", 22.254299, -100.931000}
	case "SLU":
		return Location{"Castries", "America/St_Lucia", 14.020200, -60.992901}
	case "SLV":
		return Location{"", "Asia/Kolkata", 31.081800, 77.068001}
	case "SLX":
		return Location{"Salt Cay", "America/Grand_Turk", 21.333000, -71.199997}
	case "SLY":
		return Location{"Salekhard", "Asia/Yekaterinburg", 66.590797, 66.611000}
	case "SLZ":
		return Location{"Sao Luis", "America/Fortaleza", -2.585360, -44.234100}
	case "SMA":
		return Location{"Vila do Porto", "Atlantic/Azores", 36.971401, -25.170601}
	case "SMF":
		return Location{"Sacramento", "America/Los_Angeles", 38.695400, -121.591003}
	case "SMI":
		return Location{"Samos Island", "Europe/Athens", 37.689999, 26.911699}
	case "SMK":
		return Location{"St Michael", "America/Nome", 63.490101, -162.110001}
	case "SML":
		return Location{"Stella Maris", "America/Nassau", 23.582317, -75.268621}
	case "SMQ":
		return Location{"Sampit-Borneo Island", "Asia/Pontianak", -2.499190, 112.974998}
	case "SMR":
		return Location{"Santa Marta", "America/Bogota", 11.119600, -74.230600}
	case "SMS":
		return Location{"", "Indian/Antananarivo", -17.093901, 49.815800}
	case "SMX":
		return Location{"Santa Maria", "America/Los_Angeles", 34.898899, -120.457001}
	case "SNA":
		return Location{"Santa Ana", "America/Los_Angeles", 33.675701, -117.867996}
	case "SNE":
		return Location{"Preguica", "Atlantic/Cape_Verde", 16.588400, -24.284700}
	case "SNN":
		return Location{"Shannon", "Europe/Dublin", 52.702000, -8.924820}
	case "SNO":
		return Location{"", "Asia/Bangkok", 17.195101, 104.119003}
	case "SNP":
		return Location{"St Paul Island", "America/Nome", 57.167301, -170.220001}
	case "SNR":
		return Location{"Saint-Nazaire/Montoir", "Europe/Paris", 47.312199, -2.149180}
	case "SNU":
		return Location{"Santa Clara", "America/Havana", 22.492201, -79.943604}
	case "SNW":
		return Location{"Thandwe", "Asia/Yangon", 18.460699, 94.300102}
	case "SOC":
		return Location{"Sukarata(Solo)-Java Island", "Asia/Jakarta", -7.516090, 110.757004}
	case "SOF":
		return Location{"Sofia", "Europe/Sofia", 42.696693, 23.411436}
	case "SOG":
		return Location{"Sogndal", "Europe/Oslo", 61.156101, 7.137780}
	case "SOJ":
		return Location{"Sorkjosen", "Europe/Oslo", 69.786797, 20.959400}
	case "SON":
		return Location{"Luganville", "Pacific/Efate", -15.505000, 167.220001}
	case "SOQ":
		return Location{"Sorong-Papua Island", "Asia/Jayapura", -0.926358, 131.121002}
	case "SOU":
		return Location{"Southampton", "Europe/London", 50.950298, -1.356800}
	case "SOV":
		return Location{"Seldovia", "America/Anchorage", 59.442402, -151.703995}
	case "SOW":
		return Location{"Show Low", "America/Phoenix", 34.265499, -110.005997}
	case "SPC":
		return Location{"Sta Cruz de la Palma", "Atlantic/Canary", 28.626499, -17.755600}
	case "SPD":
		return Location{"Saidpur", "Asia/Dhaka", 25.759199, 88.908897}
	case "SPI":
		return Location{"Springfield", "America/Chicago", 39.844101, -89.677902}
	case "SPN":
		return Location{"Saipan Island", "Pacific/Saipan", 15.119000, 145.729004}
	case "SPP":
		return Location{"Menongue", "Africa/Luanda", -14.657600, 17.719801}
	case "SPS":
		return Location{"Wichita Falls", "America/Chicago", 33.988800, -98.491898}
	case "SPU":
		return Location{"Split", "Europe/Zagreb", 43.538898, 16.298000}
	case "SPX":
		return Location{"Giza", "Africa/Cairo", 30.114722, 30.893333}
	case "SPY":
		return Location{"", "Africa/Abidjan", 4.746720, -6.660820}
	case "SQD":
		return Location{"San Francisco de Yeso", "America/Lima", -6.616680, -77.766701}
	case "SQG":
		return Location{"Sintang-Borneo Island", "Asia/Pontianak", 0.063619, 111.473000}
	case "SQJ":
		return Location{"Sanming", "Asia/Shanghai", 26.428056, 117.845000}
	case "SQL":
		return Location{"San Carlos", "America/Los_Angeles", 37.511902, -122.250000}
	case "SRA":
		return Location{"Santa Rosa", "America/Sao_Paulo", -27.906700, -54.520401}
	case "SRE":
		return Location{"Sucre", "America/La_Paz", -19.007099, -65.288696}
	case "SRG":
		return Location{"Semarang-Java Island", "Asia/Jakarta", -6.972730, 110.375000}
	case "SRP":
		return Location{"Svea", "Arctic/Longyearbyen", 59.791901, 5.340850}
	case "SRQ":
		return Location{"Sarasota/Bradenton", "America/New_York", 27.395399, -82.554398}
	case "SRY":
		return Location{"Sari", "Asia/Tehran", 36.635799, 53.193600}
	case "SSA":
		return Location{"Salvador", "America/Bahia", -12.908611, -38.322498}
	case "SSG":
		return Location{"Malabo", "Africa/Malabo", 3.755270, 8.708720}
	case "SSH":
		return Location{"Sharm el-Sheikh", "Africa/Cairo", 27.977301, 34.395000}
	case "SSJ":
		return Location{"Alstahaug", "Europe/Oslo", 65.956802, 12.468900}
	case "SSR":
		return Location{"Pentecost Island", "Pacific/Efate", -15.470800, 168.151993}
	case "STB":
		return Location{"", "America/Caracas", 8.974550, -71.943253}
	case "STC":
		return Location{"St Cloud", "America/Chicago", 45.546600, -94.059898}
	case "STD":
		return Location{"Santo Domingo", "America/Caracas", 7.565380, -72.035103}
	case "STI":
		return Location{"Santiago", "America/Santo_Domingo", 19.406099, -70.604698}
	case "STL":
		return Location{"St Louis", "America/Chicago", 38.748699, -90.370003}
	case "STM":
		return Location{"Santarem", "America/Santarem", -2.424722, -54.785831}
	case "STN":
		return Location{"London", "Europe/London", 51.884998, 0.235000}
	case "STR":
		return Location{"Stuttgart", "Europe/Berlin", 48.689899, 9.221960}
	case "STS":
		return Location{"Santa Rosa", "America/Los_Angeles", 38.508999, -122.813003}
	case "STT":
		return Location{"Charlotte Amalie", "America/St_Thomas", 18.337299, -64.973396}
	case "STV":
		return Location{"", "Asia/Kolkata", 21.114100, 72.741798}
	case "STW":
		return Location{"Stavropol", "Europe/Moscow", 45.109200, 42.112801}
	case "STX":
		return Location{"Christiansted", "America/St_Thomas", 17.701900, -64.798599}
	case "SUB":
		return Location{"Surabaya", "Asia/Jakarta", -7.379830, 112.787003}
	case "SUF":
		return Location{"Lamezia Terme", "Europe/Rome", 38.905399, 16.242300}
	case "SUG":
		return Location{"Surigao City", "Asia/Manila", 9.755838, 125.480947}
	case "SUJ":
		return Location{"Satu Mare", "Europe/Bucharest", 47.703300, 22.885700}
	case "SUN":
		return Location{"Hailey", "America/Boise", 43.504398, -114.295998}
	case "SUV":
		return Location{"Nausori", "Pacific/Fiji", -18.043301, 178.559006}
	case "SUX":
		return Location{"Sioux City", "America/Chicago", 42.402599, -96.384399}
	case "SUY":
		return Location{"Suntar", "Asia/Yakutsk", 62.185001, 117.635002}
	case "SVA":
		return Location{"Savoonga", "America/Nome", 63.686401, -170.492996}
	case "SVB":
		return Location{"", "Indian/Antananarivo", -14.278600, 50.174702}
	case "SVC":
		return Location{"Silver City", "America/Denver", 32.636501, -108.155998}
	case "SVD":
		return Location{"Kingstown", "America/St_Vincent", 13.144300, -61.210899}
	case "SVG":
		return Location{"Stavanger", "Europe/Oslo", 58.876701, 5.637780}
	case "SVI":
		return Location{"San Vicente Del Caguan", "America/Bogota", 2.152170, -74.766300}
	case "SVJ":
		return Location{"Svolvaer", "Europe/Oslo", 68.243301, 14.669200}
	case "SVL":
		return Location{"Savonlinna", "Europe/Helsinki", 61.943100, 28.945101}
	case "SVO":
		return Location{"Moscow", "Europe/Moscow", 55.972599, 37.414600}
	case "SVQ":
		return Location{"Sevilla", "Europe/Madrid", 37.417999, -5.893110}
	case "SVU":
		return Location{"Savusavu", "Pacific/Fiji", -16.802799, 179.341003}
	case "SVX":
		return Location{"Yekaterinburg", "Asia/Yekaterinburg", 56.743099, 60.802700}
	case "SVZ":
		return Location{"", "America/Bogota", 7.840830, -72.439697}
	case "SWA":
		return Location{"Shantou", "Asia/Shanghai", 23.426901, 116.762001}
	case "SWF":
		return Location{"Newburgh", "America/New_York", 41.504101, -74.104797}
	case "SWJ":
		return Location{"Malekula Island", "Pacific/Efate", -16.486400, 167.447200}
	case "SWO":
		return Location{"Stillwater", "America/Chicago", 36.161201, -97.085701}
	case "SWQ":
		return Location{"Sumbawa Island", "Asia/Makassar", -8.489040, 117.412003}
	case "SWT":
		return Location{"Strezhevoy", "Asia/Tomsk", 60.709400, 77.660004}
	case "SWV":
		return Location{"Evensk", "Asia/Magadan", 61.921665, 159.229996}
	case "SXB":
		return Location{"Strasbourg", "Europe/Paris", 48.538300, 7.628230}
	case "SXK":
		return Location{"Saumlaki-Yamdena Island", "Asia/Jayapura", -7.988610, 131.306000}
	case "SXM":
		return Location{"Saint Martin", "America/Lower_Princes", 18.041000, -63.108898}
	case "SXR":
		return Location{"Srinagar", "Asia/Kolkata", 33.987099, 74.774200}
	case "SXV":
		return Location{"", "Asia/Kolkata", 11.783300, 78.065598}
	case "SXZ":
		return Location{"Siirt", "Europe/Istanbul", 37.978901, 41.840401}
	case "SYD":
		return Location{"Sydney", "Australia/Sydney", -33.946098, 151.177002}
	case "SYJ":
		return Location{"", "Asia/Tehran", 29.550900, 55.672699}
	case "SYM":
		return Location{"Simao", "Asia/Shanghai", 22.793301, 100.959000}
	case "SYO":
		return Location{"Shonai", "Asia/Tokyo", 38.812199, 139.787003}
	case "SYR":
		return Location{"Syracuse", "America/New_York", 43.111198, -76.106300}
	case "SYS":
		return Location{"Saskylakh", "Asia/Yakutsk", 71.927902, 114.080002}
	case "SYU":
		return Location{"Sue Islet", "Australia/Brisbane", -10.208300, 142.824997}
	case "SYX":
		return Location{"Sanya", "Asia/Shanghai", 18.302900, 109.412003}
	case "SYY":
		return Location{"Stornoway", "Europe/London", 58.215599, -6.331110}
	case "SYZ":
		return Location{"Shiraz", "Asia/Tehran", 29.539200, 52.589802}
	case "SZA":
		return Location{"Soyo", "Africa/Luanda", -6.141090, 12.371800}
	case "SZB":
		return Location{"Subang", "Asia/Kuala_Lumpur", 3.130580, 101.549004}
	case "SZE":
		return Location{"Semera", "Africa/Addis_Ababa", 11.787500, 40.991667}
	case "SZF":
		return Location{"Samsun", "Europe/Istanbul", 41.254501, 36.567101}
	case "SZG":
		return Location{"Salzburg", "Europe/Berlin", 47.793301, 13.004300}
	case "SZK":
		return Location{"Skukuza", "Africa/Johannesburg", -24.960899, 31.588699}
	case "SZX":
		return Location{"Shenzhen", "Asia/Shanghai", 22.639299, 113.810997}
	case "SZY":
		return Location{"Szymany", "Europe/Warsaw", 53.481944, 20.937778}
	case "SZZ":
		return Location{"Goleniow", "Europe/Warsaw", 53.584702, 14.902200}
	case "TAB":
		return Location{"Scarborough", "America/Port_of_Spain", 11.149700, -60.832199}
	case "TAC":
		return Location{"Tacloban City", "Asia/Manila", 11.227600, 125.028000}
	case "TAE":
		return Location{"Daegu", "Asia/Seoul", 35.894100, 128.658997}
	case "TAG":
		return Location{"Tagbilaran City", "Asia/Manila", 9.664080, 123.852997}
	case "TAH":
		return Location{"", "Pacific/Efate", -19.455099, 169.223999}
	case "TAK":
		return Location{"Takamatsu", "Asia/Tokyo", 34.214199, 134.016006}
	case "TAL":
		return Location{"Tanana", "America/Anchorage", 65.174400, -152.108994}
	case "TAM":
		return Location{"Tampico", "America/Monterrey", 22.296400, -97.865898}
	case "TAO":
		return Location{"Qingdao", "Asia/Shanghai", 36.266102, 120.374001}
	case "TAP":
		return Location{"Tapachula", "America/Mexico_City", 14.794300, -92.370003}
	case "TAS":
		return Location{"Tashkent", "Asia/Tashkent", 41.257900, 69.281197}
	case "TAT":
		return Location{"Poprad", "Europe/Bratislava", 49.073601, 20.241100}
	case "TAY":
		return Location{"Tartu", "Europe/Tallinn", 58.307499, 26.690399}
	case "TAZ":
		return Location{"Dashoguz", "Asia/Ashgabat", 41.761101, 59.826698}
	case "TBB":
		return Location{"Tuy Hoa", "Asia/Ho_Chi_Minh", 13.049600, 109.334000}
	case "TBG":
		return Location{"Tabubil", "Pacific/Port_Moresby", -5.278610, 141.225998}
	case "TBH":
		return Location{"Romblon", "Asia/Manila", 12.311000, 122.084999}
	case "TBN":
		return Location{"Fort Leonard Wood", "America/Chicago", 37.741600, -92.140701}
	case "TBO":
		return Location{"Tabora", "Africa/Dar_es_Salaam", -5.076390, 32.833302}
	case "TBP":
		return Location{"Tumbes", "America/Lima", -3.552530, -80.381401}
	case "TBS":
		return Location{"Tbilisi", "Asia/Tbilisi", 41.669201, 44.954700}
	case "TBT":
		return Location{"Tabatinga", "America/Bogota", -4.255670, -69.935799}
	case "TBU":
		return Location{"Nuku'alofa", "Pacific/Tongatapu", -21.241199, -175.149994}
	case "TBW":
		return Location{"Tambov", "Europe/Moscow", 52.806099, 41.482800}
	case "TBZ":
		return Location{"Tabriz", "Asia/Tehran", 38.133900, 46.235001}
	case "TCA":
		return Location{"Tennant Creek", "Australia/Darwin", -19.634399, 134.182999}
	case "TCG":
		return Location{"Tacheng", "Asia/Shanghai", 46.672501, 83.340797}
	case "TCO":
		return Location{"Tumaco", "America/Bogota", 1.814420, -78.749200}
	case "TCQ":
		return Location{"Tacna", "America/Lima", -18.053301, -70.275803}
	case "TCR":
		return Location{"Thoothukkudi", "Asia/Kolkata", 8.724167, 78.025833}
	case "TCZ":
		return Location{"Tengchong", "Asia/Shanghai", 24.938056, 98.485833}
	case "TDD":
		return Location{"Trinidad", "America/La_Paz", -14.818700, -64.917999}
	case "TDK":
		return Location{"Taldy Kurgan", "Asia/Almaty", 45.126202, 78.446999}
	case "TDX":
		return Location{"", "Asia/Bangkok", 12.274600, 102.319000}
	case "TEC":
		return Location{"Telemaco Borba", "America/Sao_Paulo", -24.317801, -50.651600}
	case "TEE":
		return Location{"Tebessi", "Africa/Algiers", 35.431599, 8.120720}
	case "TEI":
		return Location{"Tezu", "Asia/Kolkata", 27.941200, 96.134399}
	case "TEN":
		return Location{"", "Asia/Shanghai", 27.883333, 109.308889}
	case "TEQ":
		return Location{"Corlu", "Europe/Istanbul", 41.138199, 27.919100}
	case "TER":
		return Location{"Lajes", "Atlantic/Azores", 38.761799, -27.090799}
	case "TET":
		return Location{"Tete", "Africa/Maputo", -16.104799, 33.640202}
	case "TEX":
		return Location{"Telluride", "America/Denver", 37.953800, -107.907997}
	case "TEZ":
		return Location{"", "Asia/Kolkata", 26.709101, 92.784698}
	case "TFF":
		return Location{"Tefe", "America/Manaus", -3.382940, -64.724098}
	case "TFL":
		return Location{"Teofilo Otoni", "America/Sao_Paulo", -17.892300, -41.513599}
	case "TFN":
		return Location{"Tenerife Island", "Atlantic/Canary", 28.482700, -16.341499}
	case "TFS":
		return Location{"Tenerife Island", "Atlantic/Canary", 28.044500, -16.572500}
	case "TFU":
		return Location{"Chengdu", "Asia/Shanghai", 30.319000, 104.445000}
	case "TGC":
		return Location{"Tanjung Manis", "Asia/Kuching", 2.177840, 111.202003}
	case "TGD":
		return Location{"Podgorica", "Europe/Podgorica", 42.359402, 19.251900}
	case "TGG":
		return Location{"Kuala Terengganu", "Asia/Kuala_Lumpur", 5.382640, 103.102997}
	case "TGI":
		return Location{"Tingo Maria", "America/Lima", -9.133000, -75.949997}
	case "TGM":
		return Location{"Targu Mures", "Europe/Bucharest", 46.467701, 24.412500}
	case "TGO":
		return Location{"Tongliao", "Asia/Shanghai", 43.556702, 122.199997}
	case "TGP":
		return Location{"Bor", "Asia/Krasnoyarsk", 61.589699, 89.994003}
	case "TGR":
		return Location{"Touggourt", "Africa/Algiers", 33.067799, 6.088670}
	case "TGT":
		return Location{"Tanga", "Africa/Dar_es_Salaam", -5.092360, 39.071201}
	case "TGU":
		return Location{"Tegucigalpa", "America/Tegucigalpa", 14.060900, -87.217201}
	case "TGZ":
		return Location{"Tuxtla Gutierrez", "America/Mexico_City", 16.563601, -93.022499}
	case "THD":
		return Location{"Thanh Hóa", "Asia/Bangkok", 19.901667, 105.467778}
	case "THE":
		return Location{"Teresina", "America/Fortaleza", -5.059940, -42.823502}
	case "THL":
		return Location{"Tachileik", "Asia/Yangon", 20.483801, 99.935402}
	case "THN":
		return Location{"Trollhattan", "Europe/Stockholm", 58.318100, 12.345000}
	case "THQ":
		return Location{"Tianshui", "Asia/Shanghai", 34.559399, 105.860001}
	case "THR":
		return Location{"Tehran", "Asia/Tehran", 35.689201, 51.313400}
	case "THS":
		return Location{"", "Asia/Bangkok", 17.238001, 99.818199}
	case "THU":
		return Location{"Thule", "America/Thule", 76.531197, -68.703201}
	case "THX":
		return Location{"Turukhansk", "Asia/Krasnoyarsk", 65.797203, 87.935303}
	case "TIA":
		return Location{"Tirana", "Europe/Tirane", 41.414700, 19.720600}
	case "TID":
		return Location{"Tiaret", "Africa/Algiers", 35.341099, 1.463150}
	case "TIF":
		return Location{"", "Asia/Riyadh", 21.483400, 40.544300}
	case "TIH":
		return Location{"", "Pacific/Tahiti", -15.119600, -148.231003}
	case "TIJ":
		return Location{"Tijuana", "America/Los_Angeles", 32.541100, -116.970001}
	case "TIM":
		return Location{"Timika-Papua Island", "Asia/Jayapura", -4.528280, 136.886993}
	case "TIN":
		return Location{"Tindouf", "Africa/Algiers", 27.700399, -8.167100}
	case "TIP":
		return Location{"Tripoli", "Africa/Tripoli", 32.663502, 13.159000}
	case "TIR":
		return Location{"Tirupati", "Asia/Kolkata", 13.632500, 79.543297}
	case "TIU":
		return Location{"", "Pacific/Auckland", -44.302799, 171.225006}
	case "TIV":
		return Location{"Tivat", "Europe/Podgorica", 42.404701, 18.723301}
	case "TIZ":
		return Location{"Tari", "Pacific/Port_Moresby", -5.845000, 142.947998}
	case "TJA":
		return Location{"Tarija", "America/La_Paz", -21.555700, -64.701302}
	case "TJH":
		return Location{"Tajima", "Asia/Tokyo", 35.512798, 134.787003}
	case "TJK":
		return Location{"Tokat", "Europe/Istanbul", 40.307430, 36.367409}
	case "TJL":
		return Location{"Tres Lagoas", "America/Campo_Grande", -20.754444, -51.683889}
	case "TJM":
		return Location{"Tyumen", "Asia/Yekaterinburg", 57.189602, 65.324303}
	case "TJN":
		return Location{"Takume", "Pacific/Tahiti", -15.854700, -142.268005}
	case "TJQ":
		return Location{"Tanjung Pandan-Belitung Island", "Asia/Jakarta", -2.745720, 107.754997}
	case "TJS":
		return Location{"Tanjung Selor-Borneo Island", "Asia/Makassar", 2.836410, 117.374001}
	case "TJU":
		return Location{"Kulyab", "Asia/Dushanbe", 37.988098, 69.805000}
	case "TKD":
		return Location{"Sekondi-Takoradi", "Africa/Accra", 4.896060, -1.774760}
	case "TKF":
		return Location{"Truckee", "America/Los_Angeles", 39.320000, -120.139999}
	case "TKG":
		return Location{"Bandar Lampung-Sumatra Island", "Asia/Jakarta", -5.240556, 105.175556}
	case "TKJ":
		return Location{"Tok", "America/Anchorage", 63.303333, -143.001111}
	case "TKK":
		return Location{"Weno Island", "Pacific/Chuuk", 7.461870, 151.843002}
	case "TKN":
		return Location{"Tokunoshima", "Asia/Tokyo", 27.836399, 128.880997}
	case "TKP":
		return Location{"", "Pacific/Tahiti", -14.709500, -145.246002}
	case "TKQ":
		return Location{"Kigoma", "Africa/Dar_es_Salaam", -4.886200, 29.670900}
	case "TKS":
		return Location{"Tokushima", "Asia/Tokyo", 34.132801, 134.606995}
	case "TKU":
		return Location{"Turku", "Europe/Helsinki", 60.514099, 22.262800}
	case "TKV":
		return Location{"Tatakoto", "Pacific/Tahiti", -17.355301, -138.445007}
	case "TKX":
		return Location{"", "Pacific/Tahiti", -14.455800, -145.024994}
	case "TLA":
		return Location{"Teller", "America/Nome", 65.240402, -166.339005}
	case "TLC":
		return Location{"Toluca", "America/Mexico_City", 19.337099, -99.566002}
	case "TLE":
		return Location{"", "Indian/Antananarivo", -23.383400, 43.728500}
	case "TLH":
		return Location{"Tallahassee", "America/New_York", 30.396500, -84.350304}
	case "TLI":
		return Location{"Toli Toli-Celebes Island", "Asia/Makassar", -1.029770, 120.817001}
	case "TLK":
		return Location{"", "Asia/Yakutsk", 59.876389, 111.044444}
	case "TLL":
		return Location{"Tallinn", "Europe/Tallinn", 59.413300, 24.832800}
	case "TLM":
		return Location{"Tlemcen", "Africa/Algiers", 35.016701, -1.450000}
	case "TLN":
		return Location{"Toulon/Hyeres/Le Palyvestre", "Europe/Paris", 43.097301, 6.146030}
	case "TLS":
		return Location{"Toulouse/Blagnac", "Europe/Paris", 43.629101, 1.363820}
	case "TLU":
		return Location{"Tolu", "America/Bogota", 9.509450, -75.585400}
	case "TLV":
		return Location{"Tel Aviv", "Asia/Jerusalem", 32.011398, 34.886700}
	case "TLY":
		return Location{"Plastun", "Asia/Vladivostok", 44.814999, 136.292007}
	case "TMC":
		return Location{"Waikabubak-Sumba Island", "Asia/Makassar", -9.409720, 119.244003}
	case "TME":
		return Location{"Tame", "America/Bogota", 6.451080, -71.760300}
	case "TMI":
		return Location{"Tumling Tar", "Asia/Kathmandu", 27.315001, 87.193298}
	case "TMJ":
		return Location{"Termez", "Asia/Samarkand", 37.286701, 67.309998}
	case "TML":
		return Location{"Tamale", "Africa/Accra", 9.557190, -0.863214}
	case "TMM":
		return Location{"", "Indian/Antananarivo", -18.109501, 49.392502}
	case "TMP":
		return Location{"Tampere / Pirkkala", "Europe/Helsinki", 61.414101, 23.604401}
	case "TMR":
		return Location{"Tamanrasset", "Africa/Algiers", 22.811501, 5.451080}
	case "TMS":
		return Location{"Sao Tome", "Africa/Sao_Tome", 0.378175, 6.712150}
	case "TMT":
		return Location{"Oriximina", "America/Santarem", -1.489600, -56.396801}
	case "TMW":
		return Location{"Tamworth", "Australia/Sydney", -31.083900, 150.847000}
	case "TMX":
		return Location{"Timimoun", "Africa/Algiers", 29.237101, 0.276033}
	case "TNA":
		return Location{"Jinan", "Asia/Shanghai", 36.857201, 117.216003}
	case "TNC":
		return Location{"Tin City", "America/Nome", 65.563103, -167.921997}
	case "TNE":
		return Location{"", "Asia/Tokyo", 30.605101, 130.990997}
	case "TNG":
		return Location{"Tangier", "Africa/Casablanca", 35.726898, -5.916890}
	case "TNH":
		return Location{"Tonghua", "Asia/Shanghai", 42.253889, 125.703333}
	case "TNJ":
		return Location{"Tanjung Pinang-Bintan Island", "Asia/Jakarta", 0.922683, 104.531998}
	case "TNN":
		return Location{"Tainan City", "Asia/Taipei", 22.950399, 120.206001}
	case "TNO":
		return Location{"Santa Cruz", "America/Costa_Rica", 10.313500, -85.815498}
	case "TNR":
		return Location{"Antananarivo", "Indian/Antananarivo", -18.796900, 47.478802}
	case "TOB":
		return Location{"Tobruk", "Africa/Tripoli", 31.861000, 23.907000}
	case "TOE":
		return Location{"Tozeur", "Africa/Tunis", 33.939701, 8.110560}
	case "TOF":
		return Location{"Tomsk", "Asia/Tomsk", 56.380299, 85.208298}
	case "TOG":
		return Location{"Togiak Village", "America/Anchorage", 59.052799, -160.397003}
	case "TOH":
		return Location{"Loh/Linua", "Pacific/Efate", -13.328000, 166.638000}
	case "TOL":
		return Location{"Toledo", "America/New_York", 41.586800, -83.807800}
	case "TOS":
		return Location{"Tromso", "Europe/Oslo", 69.683296, 18.918900}
	case "TOY":
		return Location{"Toyama", "Asia/Tokyo", 36.648300, 137.188004}
	case "TPA":
		return Location{"Tampa", "America/New_York", 27.975500, -82.533203}
	case "TPE":
		return Location{"Taipei", "Asia/Taipei", 25.077700, 121.233002}
	case "TPP":
		return Location{"Tarapoto", "America/Lima", -6.508740, -76.373199}
	case "TPQ":
		return Location{"Tepic", "America/Mazatlan", 21.419500, -104.843002}
	case "TPS":
		return Location{"Trapani", "Europe/Rome", 37.911400, 12.488000}
	case "TRC":
		return Location{"Torreon", "America/Monterrey", 25.568300, -103.411003}
	case "TRD":
		return Location{"Trondheim", "Europe/Oslo", 63.457802, 10.924000}
	case "TRE":
		return Location{"Balemartine", "Europe/London", 56.499199, -6.869170}
	case "TRF":
		return Location{"Torp", "Europe/Oslo", 59.186699, 10.258600}
	case "TRG":
		return Location{"Tauranga", "Pacific/Auckland", -37.671902, 176.195999}
	case "TRI":
		return Location{"Bristol/Johnson/Kingsport", "America/New_York", 36.475201, -82.407402}
	case "TRK":
		return Location{"Tarakan Island", "Asia/Makassar", 3.326690, 117.566002}
	case "TRN":
		return Location{"Torino", "Europe/Rome", 45.200802, 7.649630}
	case "TRR":
		return Location{"Trincomalee", "Asia/Colombo", 8.538510, 81.181900}
	case "TRS":
		return Location{"Trieste", "Europe/Rome", 45.827499, 13.472200}
	case "TRU":
		return Location{"Trujillo", "America/Lima", -8.081410, -79.108803}
	case "TRV":
		return Location{"Trivandrum", "Asia/Kolkata", 8.482120, 76.920097}
	case "TRW":
		return Location{"Tarawa", "Pacific/Tarawa", 1.381640, 173.147003}
	case "TRZ":
		return Location{"Tiruchirappally", "Asia/Kolkata", 10.765400, 78.709702}
	case "TSA":
		return Location{"Taipei City", "Asia/Taipei", 25.069401, 121.552002}
	case "TSF":
		return Location{"Treviso", "Europe/Rome", 45.648399, 12.194400}
	case "TSJ":
		return Location{"Tsushima", "Asia/Tokyo", 34.284901, 129.330994}
	case "TSM":
		return Location{"Taos", "America/Denver", 36.458199, -105.671997}
	case "TSN":
		return Location{"Tianjin", "Asia/Shanghai", 39.124401, 117.346001}
	case "TSR":
		return Location{"Timisoara", "Europe/Bucharest", 45.809898, 21.337900}
	case "TST":
		return Location{"", "Asia/Bangkok", 7.508740, 99.616600}
	case "TSV":
		return Location{"Townsville", "Australia/Brisbane", -19.252501, 146.764999}
	case "TTA":
		return Location{"Tan Tan", "Africa/Casablanca", 28.448200, -11.161300}
	case "TTE":
		return Location{"Sango-Ternate Island", "Asia/Jayapura", 0.831414, 127.380997}
	case "TTJ":
		return Location{"Tottori", "Asia/Tokyo", 35.530102, 134.167007}
	case "TTN":
		return Location{"Trenton", "America/New_York", 40.276699, -74.813499}
	case "TTQ":
		return Location{"Roxana", "America/Costa_Rica", 10.569000, -83.514801}
	case "TTT":
		return Location{"Taitung City", "Asia/Taipei", 22.754999, 121.101997}
	case "TTU":
		return Location{"", "Africa/Casablanca", 35.594299, -5.320020}
	case "TUB":
		return Location{"", "Pacific/Tahiti", -23.365400, -149.524002}
	case "TUC":
		return Location{"San Miguel de Tucuman", "America/Argentina/Tucuman", -26.840900, -65.104897}
	case "TUF":
		return Location{"Tours/Val de Loire (Loire Valley)", "Europe/Paris", 47.432201, 0.727606}
	case "TUG":
		return Location{"Tuguegarao City", "Asia/Manila", 17.643368, 121.733150}
	case "TUI":
		return Location{"", "Asia/Riyadh", 31.692699, 38.731201}
	case "TUK":
		return Location{"Turbat", "Asia/Karachi", 25.986401, 63.030201}
	case "TUL":
		return Location{"Tulsa", "America/Chicago", 36.198399, -95.888100}
	case "TUN":
		return Location{"Tunis", "Africa/Tunis", 36.851002, 10.227200}
	case "TUO":
		return Location{"Taupo", "Pacific/Auckland", -38.739700, 176.084000}
	case "TUP":
		return Location{"Tupelo", "America/Chicago", 34.268101, -88.769897}
	case "TUR":
		return Location{"Tucurui", "America/Belem", -3.786010, -49.720299}
	case "TUS":
		return Location{"Tucson", "America/Phoenix", 32.116100, -110.941002}
	case "TUU":
		return Location{"", "Asia/Riyadh", 28.365400, 36.618900}
	case "TVC":
		return Location{"Traverse City", "America/Detroit", 44.741402, -85.582199}
	case "TVF":
		return Location{"Thief River Falls", "America/Chicago", 48.065701, -96.184998}
	case "TVU":
		return Location{"Matei", "Pacific/Fiji", -16.690599, -179.876999}
	case "TVY":
		return Location{"Dawei", "Asia/Yangon", 14.103900, 98.203598}
	case "TWF":
		return Location{"Twin Falls", "America/Boise", 42.481800, -114.487999}
	case "TWU":
		return Location{"Tawau", "Asia/Kuching", 4.320160, 118.127998}
	case "TXF":
		return Location{"Teixeira De Freitas", "America/Bahia", -17.524500, -39.668499}
	case "TXK":
		return Location{"Texarkana", "America/Chicago", 33.453701, -93.990997}
	case "TXN":
		return Location{"Huangshan", "Asia/Shanghai", 29.733299, 118.255997}
	case "TYD":
		return Location{"Tynda", "Asia/Yakutsk", 55.284199, 124.778999}
	case "TYF":
		return Location{"", "Europe/Stockholm", 60.157600, 12.991300}
	case "TYL":
		return Location{"", "America/Lima", -4.576640, -81.254097}
	case "TYN":
		return Location{"Taiyuan", "Asia/Shanghai", 37.746899, 112.627998}
	case "TYR":
		return Location{"Tyler", "America/Chicago", 32.354099, -95.402397}
	case "TYS":
		return Location{"Knoxville", "America/New_York", 35.811001, -83.994003}
	case "TZL":
		return Location{"Tuzla", "Europe/Sarajevo", 44.458698, 18.724800}
	case "TZX":
		return Location{"Trabzon", "Europe/Istanbul", 40.995098, 39.789700}
	case "UAH":
		return Location{"Ua Huka", "Pacific/Marquesas", -8.936110, -139.552002}
	case "UAK":
		return Location{"Narsarsuaq", "America/Nuuk", 61.160500, -45.425999}
	case "UAP":
		return Location{"Ua Pou", "Pacific/Marquesas", -9.351670, -140.078003}
	case "UAQ":
		return Location{"San Juan", "America/Argentina/San_Juan", -31.571501, -68.418198}
	case "UAS":
		return Location{"Samburu South", "Africa/Nairobi", 0.530583, 37.534195}
	case "UBA":
		return Location{"Uberaba", "America/Sao_Paulo", -19.764723, -47.966110}
	case "UBB":
		return Location{"Mabuiag Island", "Australia/Brisbane", -9.950000, 142.182999}
	case "UBJ":
		return Location{"Ube", "Asia/Tokyo", 33.930000, 131.279007}
	case "UBN":
		return Location{"Ulaanbaatar", "Asia/Ulaanbaatar", 47.651581, 106.821772}
	case "UBP":
		return Location{"Ubon Ratchathani", "Asia/Bangkok", 15.251300, 104.870003}
	case "UCT":
		return Location{"Ukhta", "Europe/Moscow", 63.566898, 53.804699}
	case "UDI":
		return Location{"Uberlandia", "America/Sao_Paulo", -18.883612, -48.225277}
	case "UDR":
		return Location{"Udaipur", "Asia/Kolkata", 24.617701, 73.896103}
	case "UEL":
		return Location{"Quelimane", "Africa/Maputo", -17.855499, 36.869099}
	case "UEO":
		return Location{"", "Asia/Tokyo", 26.363501, 126.713997}
	case "UET":
		return Location{"Quetta", "Asia/Karachi", 30.251400, 66.937798}
	case "UFA":
		return Location{"Ufa", "Asia/Yekaterinburg", 54.557499, 55.874401}
	case "UGC":
		return Location{"Urgench", "Asia/Samarkand", 41.584301, 60.641701}
	case "UIB":
		return Location{"Quibdo", "America/Bogota", 5.690760, -76.641200}
	case "UIH":
		return Location{"Quy Nohn", "Asia/Ho_Chi_Minh", 13.955000, 109.042000}
	case "UII":
		return Location{"Utila Island", "America/Tegucigalpa", 16.113100, -86.880302}
	case "UIN":
		return Location{"Quincy", "America/Chicago", 39.942699, -91.194603}
	case "UIO":
		return Location{"Quito", "America/Guayaquil", -0.129167, -78.357500}
	case "UKA":
		return Location{"Ukunda", "Africa/Nairobi", -4.293330, 39.571098}
	case "UKB":
		return Location{"Kobe", "Asia/Tokyo", 34.632801, 135.223999}
	case "UKG":
		return Location{"Ust-Kuyga", "Asia/Vladivostok", 70.011002, 135.645004}
	case "UKK":
		return Location{"Ust Kamenogorsk", "Asia/Almaty", 50.036598, 82.494202}
	case "UKX":
		return Location{"Ust-Kut", "Asia/Irkutsk", 56.856701, 105.730003}
	case "ULB":
		return Location{"Ambryn Island", "Pacific/Efate", -16.329700, 168.301100}
	case "ULG":
		return Location{"", "Asia/Hovd", 48.993301, 89.922501}
	case "ULH":
		return Location{"Al-'Ula", "Asia/Riyadh", 26.483333, 38.116944}
	case "ULK":
		return Location{"Lensk", "Asia/Yakutsk", 60.720600, 114.825996}
	case "ULO":
		return Location{"", "Asia/Hovd", 49.973333, 92.079722}
	case "ULP":
		return Location{"", "Australia/Brisbane", -26.612200, 144.253006}
	case "ULV":
		return Location{"Ulyanovsk", "Europe/Ulyanovsk", 54.268299, 48.226700}
	case "UME":
		return Location{"Umea", "Europe/Stockholm", 63.791801, 20.282801}
	case "UMU":
		return Location{"Umuarama", "America/Sao_Paulo", -23.798700, -53.313801}
	case "UNA":
		return Location{"Una", "America/Bahia", -15.355200, -38.999001}
	case "UNG":
		return Location{"Kiunga", "Pacific/Port_Moresby", -6.125710, 141.281998}
	case "UNK":
		return Location{"Unalakleet", "America/Anchorage", 63.888401, -160.798996}
	case "UNN":
		return Location{"", "Asia/Bangkok", 9.777620, 98.585503}
	case "UPG":
		return Location{"Ujung Pandang-Celebes Island", "Asia/Makassar", -5.061630, 119.554001}
	case "UPN":
		return Location{"", "America/Mexico_City", 19.396700, -102.039001}
	case "URA":
		return Location{"Uralsk", "Asia/Oral", 51.150799, 51.543098}
	case "URC":
		return Location{"Urumqi", "Asia/Shanghai", 43.907101, 87.474197}
	case "URE":
		return Location{"Kuressaare", "Europe/Tallinn", 58.229900, 22.509501}
	case "URG":
		return Location{"Uruguaiana", "America/Sao_Paulo", -29.782200, -57.038200}
	case "URJ":
		return Location{"Uray", "Asia/Yekaterinburg", 60.103298, 64.826698}
	case "URT":
		return Location{"Surat Thani", "Asia/Bangkok", 9.132600, 99.135597}
	case "URY":
		return Location{"", "Asia/Riyadh", 31.411900, 37.279499}
	case "USH":
		return Location{"Ushuahia", "America/Argentina/Ushuaia", -54.843300, -68.295800}
	case "USK":
		return Location{"Usinsk", "Europe/Moscow", 66.004700, 57.367199}
	case "USM":
		return Location{"Na Thon (Ko Samui Island)", "Asia/Bangkok", 9.547790, 100.061996}
	case "USN":
		return Location{"Ulsan", "Asia/Seoul", 35.593498, 129.352005}
	case "USR":
		return Location{"Ust-Nera", "Asia/Ust-Nera", 64.550003, 143.115005}
	case "USU":
		return Location{"Coron", "Asia/Manila", 12.121500, 120.099998}
	case "UTH":
		return Location{"Udon Thani", "Asia/Bangkok", 17.386400, 102.788002}
	case "UTN":
		return Location{"Upington", "Africa/Johannesburg", -28.399099, 21.260201}
	case "UTP":
		return Location{"Rayong", "Asia/Bangkok", 12.679900, 101.004997}
	case "UTT":
		return Location{"Mthatha", "Africa/Johannesburg", -31.547899, 28.674299}
	case "UUA":
		return Location{"Bugulma", "Europe/Moscow", 54.639999, 52.801701}
	case "UUD":
		return Location{"Ulan Ude", "Asia/Irkutsk", 51.807800, 107.438004}
	case "UUS":
		return Location{"Yuzhno-Sakhalinsk", "Asia/Sakhalin", 46.888699, 142.718002}
	case "UVE":
		return Location{"Ouvea", "Pacific/Noumea", -20.640600, 166.572998}
	case "UVF":
		return Location{"Vieux Fort", "America/St_Lucia", 13.733200, -60.952599}
	case "UYN":
		return Location{"Yulin", "Asia/Shanghai", 38.269199, 109.731003}
	case "UYU":
		return Location{"Quijarro", "America/La_Paz", -20.446301, -66.848396}
	case "VAA":
		return Location{"Vaasa", "Europe/Helsinki", 63.050701, 21.762199}
	case "VAG":
		return Location{"Varginha", "America/Sao_Paulo", -21.590099, -45.473301}
	case "VAI":
		return Location{"", "Pacific/Port_Moresby", -2.697170, 141.302002}
	case "VAK":
		return Location{"Chevak", "America/Nome", 61.540900, -165.600500}
	case "VAM":
		return Location{"Maamigili", "Indian/Maldives", 3.470556, 72.835833}
	case "VAN":
		return Location{"Van", "Europe/Istanbul", 38.468201, 43.332298}
	case "VAO":
		return Location{"Suavanao", "Pacific/Guadalcanal", -7.585560, 158.731003}
	case "VAR":
		return Location{"Varna", "Europe/Sofia", 43.232101, 27.825100}
	case "VAS":
		return Location{"Sivas", "Europe/Istanbul", 39.813801, 36.903500}
	case "VAV":
		return Location{"Vava'u Island", "Pacific/Tongatapu", -18.585300, -173.962006}
	case "VAW":
		return Location{"Vardo", "Europe/Oslo", 70.355400, 31.044901}
	case "VBA":
		return Location{"Aeng", "Asia/Yangon", 19.769199, 94.026100}
	case "VBP":
		return Location{"Bokpyinn", "Asia/Yangon", 11.267000, 98.766998}
	case "VBV":
		return Location{"Vanua Balavu", "Pacific/Fiji", -17.268999, -178.975998}
	case "VBY":
		return Location{"Visby", "Europe/Stockholm", 57.662800, 18.346201}
	case "VCA":
		return Location{"Can Tho", "Asia/Ho_Chi_Minh", 10.085100, 105.711998}
	case "VCE":
		return Location{"Venezia", "Europe/Rome", 45.505299, 12.351900}
	case "VCL":
		return Location{"Dung Quat Bay", "Asia/Ho_Chi_Minh", 15.403300, 108.706001}
	case "VCP":
		return Location{"Campinas", "America/Sao_Paulo", -23.007401, -47.134499}
	case "VCS":
		return Location{"Con Ong", "Asia/Ho_Chi_Minh", 8.731830, 106.633003}
	case "VCT":
		return Location{"Victoria", "America/Chicago", 28.852600, -96.918503}
	case "VDC":
		return Location{"Vitoria Da Conquista", "America/Bahia", -14.862800, -40.863098}
	case "VDE":
		return Location{"El Hierro Island", "Atlantic/Canary", 27.814800, -17.887100}
	case "VDH":
		return Location{"Dong Hoi", "Asia/Bangkok", 17.515000, 106.590556}
	case "VDM":
		return Location{"Viedma / Carmen de Patagones", "America/Argentina/Salta", -40.869200, -63.000400}
	case "VDS":
		return Location{"Vadso", "Europe/Oslo", 70.065300, 29.844700}
	case "VDY":
		return Location{"", "Asia/Kolkata", 15.174967, 76.634947}
	case "VDZ":
		return Location{"Valdez", "America/Anchorage", 61.133900, -146.248001}
	case "VEE":
		return Location{"Venetie", "America/Anchorage", 67.008698, -146.365997}
	case "VEL":
		return Location{"Vernal", "America/Denver", 40.440899, -109.510002}
	case "VER":
		return Location{"Veracruz", "America/Mexico_City", 19.145901, -96.187302}
	case "VFA":
		return Location{"Victoria Falls", "Africa/Harare", -18.095900, 25.839001}
	case "VGA":
		return Location{"", "Asia/Kolkata", 16.530399, 80.796799}
	case "VGO":
		return Location{"Vigo", "Europe/Madrid", 42.231800, -8.626770}
	case "VGZ":
		return Location{"Villa Garzon", "America/Bogota", 0.981944, -76.604167}
	case "VHC":
		return Location{"Saurimo", "Africa/Luanda", -9.689070, 20.431900}
	case "VHM":
		return Location{"", "Europe/Stockholm", 64.579102, 16.833599}
	case "VHZ":
		return Location{"Vahitahi", "Pacific/Tahiti", -18.780001, -138.852997}
	case "VIE":
		return Location{"Vienna", "Europe/Vienna", 48.110298, 16.569700}
	case "VIG":
		return Location{"El Vigia", "America/Caracas", 8.624139, -71.672668}
	case "VII":
		return Location{"Vinh", "Asia/Bangkok", 18.737600, 105.670998}
	case "VIJ":
		return Location{"Spanish Town", "America/Tortola", 18.446400, -64.427498}
	case "VIL":
		return Location{"Dakhla", "Africa/El_Aaiun", 23.718300, -15.932000}
	case "VIT":
		return Location{"Alava", "Europe/Madrid", 42.882801, -2.724470}
	case "VIX":
		return Location{"Vitoria", "America/Sao_Paulo", -20.258057, -40.286388}
	case "VJB":
		return Location{"Xai-Xai", "Africa/Maputo", -25.037800, 33.627399}
	case "VKG":
		return Location{"Rach Gia", "Asia/Ho_Chi_Minh", 9.958030, 105.132380}
	case "VKO":
		return Location{"Moscow", "Europe/Moscow", 55.591499, 37.261501}
	case "VKT":
		return Location{"Vorkuta", "Europe/Moscow", 67.488602, 63.993099}
	case "VLC":
		return Location{"Valencia", "Europe/Madrid", 39.489300, -0.481625}
	case "VLD":
		return Location{"Valdosta", "America/New_York", 30.782499, -83.276703}
	case "VLI":
		return Location{"Port Vila", "Pacific/Efate", -17.699301, 168.320007}
	case "VLL":
		return Location{"Valladolid", "Europe/Madrid", 41.706100, -4.851940}
	case "VLN":
		return Location{"Valencia", "America/Caracas", 10.149733, -67.928398}
	case "VLS":
		return Location{"Valesdir", "Pacific/Efate", -16.796101, 168.177002}
	case "VLV":
		return Location{"Valera", "America/Caracas", 9.340478, -70.584061}
	case "VNO":
		return Location{"Vilnius", "Europe/Vilnius", 54.634102, 25.285801}
	case "VNS":
		return Location{"Varanasi", "Asia/Kolkata", 25.452400, 82.859299}
	case "VNX":
		return Location{"Vilanculo", "Africa/Maputo", -22.018400, 35.313301}
	case "VNY":
		return Location{"Van Nuys", "America/Los_Angeles", 34.209801, -118.489998}
	case "VOG":
		return Location{"Volgograd", "Europe/Volgograd", 48.782501, 44.345501}
	case "VOL":
		return Location{"Nea Anchialos", "Europe/Athens", 39.219601, 22.794300}
	case "VPE":
		return Location{"Ngiva", "Africa/Luanda", -17.043501, 15.683800}
	case "VPS":
		return Location{"Valparaiso", "America/Chicago", 30.483200, -86.525398}
	case "VPY":
		return Location{"Chimoio", "Africa/Maputo", -19.151300, 33.429001}
	case "VQS":
		return Location{"Vieques Island", "America/Puerto_Rico", 18.134800, -65.493599}
	case "VRA":
		return Location{"Varadero", "America/Havana", 23.034401, -81.435303}
	case "VRB":
		return Location{"Vero Beach", "America/New_York", 27.655600, -80.417900}
	case "VRC":
		return Location{"Virac", "Asia/Manila", 13.576400, 124.206001}
	case "VRN":
		return Location{"Verona", "Europe/Rome", 45.395699, 10.888500}
	case "VSA":
		return Location{"Villahermosa", "America/Mexico_City", 17.997000, -92.817398}
	case "VST":
		return Location{"Stockholm / Vasteras", "Europe/Stockholm", 59.589401, 16.633600}
	case "VTE":
		return Location{"Vientiane", "Asia/Vientiane", 17.988300, 102.563004}
	case "VTZ":
		return Location{"Visakhapatnam", "Asia/Kolkata", 17.721201, 83.224503}
	case "VUP":
		return Location{"Valledupar", "America/Bogota", 10.435000, -73.249500}
	case "VUS":
		return Location{"Velikiy Ustyug", "Europe/Moscow", 60.788300, 46.259998}
	case "VVC":
		return Location{"Villavicencio", "America/Bogota", 4.167870, -73.613800}
	case "VVI":
		return Location{"Santa Cruz", "America/La_Paz", -17.644800, -63.135399}
	case "VVO":
		return Location{"Vladivostok", "Asia/Vladivostok", 43.398998, 132.147995}
	case "VVZ":
		return Location{"Illizi", "Africa/Algiers", 26.723499, 8.622650}
	case "VXC":
		return Location{"Lichinga", "Africa/Maputo", -13.274000, 35.266300}
	case "VXE":
		return Location{"Sao Pedro", "Atlantic/Cape_Verde", 16.833200, -25.055300}
	case "VXO":
		return Location{"Vaxjo", "Europe/Stockholm", 56.929100, 14.728000}
	case "WAA":
		return Location{"Wales", "America/Nome", 65.622597, -168.095001}
	case "WAE":
		return Location{"", "Asia/Riyadh", 20.504299, 45.199600}
	case "WAG":
		return Location{"Wanganui", "Pacific/Auckland", -39.962200, 175.024994}
	case "WAW":
		return Location{"Warsaw", "Europe/Warsaw", 52.165699, 20.967100}
	case "WBM":
		return Location{"", "Pacific/Port_Moresby", -5.643300, 143.895004}
	case "WBQ":
		return Location{"Beaver", "America/Anchorage", 66.362198, -147.406998}
	case "WDH":
		return Location{"Windhoek", "Africa/Windhoek", -22.479900, 17.470900}
	case "WEF":
		return Location{"Weifang", "Asia/Shanghai", 36.646702, 119.119003}
	case "WEH":
		return Location{"Weihai", "Asia/Shanghai", 37.187099, 122.228996}
	case "WEI":
		return Location{"Weipa", "Australia/Brisbane", -12.678600, 141.925003}
	case "WGA":
		return Location{"Wagga Wagga", "Australia/Sydney", -35.165298, 147.466003}
	case "WGE":
		return Location{"", "Australia/Sydney", -30.032801, 148.126007}
	case "WGP":
		return Location{"Waingapu-Sumba Island", "Asia/Makassar", -9.669220, 120.302002}
	case "WHK":
		return Location{"", "Pacific/Auckland", -37.920601, 176.914001}
	case "WIC":
		return Location{"Wick", "Europe/London", 58.458900, -3.093060}
	case "WIL":
		return Location{"Nairobi", "Africa/Nairobi", -1.321720, 36.814800}
	case "WIN":
		return Location{"", "Australia/Brisbane", -22.363600, 143.085999}
	case "WJU":
		return Location{"Wonju", "Asia/Seoul", 37.438099, 127.959999}
	case "WKA":
		return Location{"", "Pacific/Auckland", -44.722198, 169.246002}
	case "WKJ":
		return Location{"Wakkanai", "Asia/Tokyo", 45.404202, 141.800995}
	case "WLE":
		return Location{"", "Australia/Brisbane", -26.808300, 150.175003}
	case "WLG":
		return Location{"Wellington", "Pacific/Auckland", -41.327202, 174.804993}
	case "WLH":
		return Location{"Walaha", "Pacific/Efate", -15.412000, 167.690994}
	case "WLK":
		return Location{"Selawik", "America/Anchorage", 66.600098, -159.985992}
	case "WLS":
		return Location{"Wallis Island", "Pacific/Wallis", -13.238300, -176.199005}
	case "WMI":
		return Location{"Warsaw", "Europe/Warsaw", 52.451099, 20.651800}
	case "WMN":
		return Location{"", "Indian/Antananarivo", -15.436700, 49.688301}
	case "WMO":
		return Location{"White Mountain", "America/Nome", 64.689201, -163.412994}
	case "WMX":
		return Location{"Wamena-Papua Island", "Asia/Jayapura", -4.102510, 138.957001}
	case "WNA":
		return Location{"Napakiak", "America/Anchorage", 60.690300, -161.979004}
	case "WNP":
		return Location{"Naga", "Asia/Manila", 13.584900, 123.269997}
	case "WNR":
		return Location{"", "Australia/Brisbane", -25.413099, 142.667007}
	case "WNZ":
		return Location{"Wenzhou", "Asia/Shanghai", 27.912201, 120.851997}
	case "WOL":
		return Location{"", "Australia/Sydney", -34.561100, 150.789001}
	case "WPR":
		return Location{"Porvenir", "America/Punta_Arenas", -53.253700, -70.319199}
	case "WPU":
		return Location{"Puerto Williams", "America/Argentina/Ushuaia", -54.931099, -67.626297}
	case "WRE":
		return Location{"", "Pacific/Auckland", -35.768299, 174.365005}
	case "WRG":
		return Location{"Wrangell", "America/Sitka", 56.484299, -132.369995}
	case "WRO":
		return Location{"Wroclaw", "Europe/Warsaw", 51.102699, 16.885799}
	case "WRZ":
		return Location{"Weerawila", "Asia/Colombo", 6.254490, 81.235199}
	case "WSN":
		return Location{"South Naknek", "America/Anchorage", 58.703400, -157.007996}
	case "WSZ":
		return Location{"", "Pacific/Auckland", -41.738098, 171.580994}
	case "WTB":
		return Location{"Wellcamp", "Australia/Brisbane", -27.558333, 151.793333}
	case "WTK":
		return Location{"Noatak", "America/Nome", 67.566101, -162.975006}
	case "WUH":
		return Location{"Wuhan", "Asia/Shanghai", 30.783800, 114.208000}
	case "WUN":
		return Location{"", "Australia/Perth", -26.629200, 120.221001}
	case "WUS":
		return Location{"Wuyishan", "Asia/Shanghai", 27.701900, 118.000999}
	case "WUX":
		return Location{"Wuxi", "Asia/Shanghai", 31.494400, 120.429001}
	case "WUZ":
		return Location{"Wuzhou", "Asia/Shanghai", 23.456699, 111.248001}
	case "WVB":
		return Location{"Walvis Bay", "Africa/Windhoek", -22.979900, 14.645300}
	case "WWK":
		return Location{"Wewak", "Pacific/Port_Moresby", -3.583830, 143.669006}
	case "WWT":
		return Location{"Newtok", "America/Nome", 60.939098, -164.641006}
	case "WXN":
		return Location{"Wanxian", "Asia/Shanghai", 30.801700, 108.433000}
	case "WYA":
		return Location{"Whyalla", "Australia/Adelaide", -33.058899, 137.514008}
	case "WYS":
		return Location{"West Yellowstone", "America/Denver", 44.688400, -111.117996}
	case "XAP":
		return Location{"Chapeco", "America/Sao_Paulo", -27.134199, -52.656601}
	case "XBE":
		return Location{"Bearskin Lake", "America/Rainy_River", 53.965599, -91.027199}
	case "XCH":
		return Location{"Christmas Island", "Indian/Christmas", -10.450600, 105.690002}
	case "XCR":
		return Location{"Chalons/Vatry", "Europe/Paris", 48.776100, 4.184490}
	case "XFN":
		return Location{"Xiangfan", "Asia/Shanghai", 32.150600, 112.291000}
	case "XGR":
		return Location{"Kangiqsualujjuaq", "America/Toronto", 58.711399, -65.992798}
	case "XIC":
		return Location{"Xichang", "Asia/Shanghai", 27.989100, 102.183998}
	case "XIL":
		return Location{"Xilinhot", "Asia/Shanghai", 43.915600, 115.963997}
	case "XIY":
		return Location{"Xianyang", "Asia/Shanghai", 34.447102, 108.751999}
	case "XKH":
		return Location{"Xieng Khouang", "Asia/Vientiane", 19.450001, 103.157997}
	case "XLS":
		return Location{"Saint Louis", "Africa/Dakar", 16.050800, -16.463200}
	case "XMH":
		return Location{"", "Pacific/Tahiti", -14.436800, -146.070007}
	case "XMN":
		return Location{"Xiamen", "Asia/Shanghai", 24.544001, 118.127998}
	case "XMY":
		return Location{"Yam Island", "Australia/Brisbane", -9.901110, 142.776001}
	case "XNA":
		return Location{"Fayetteville/Springdale/", "America/Chicago", 36.281898, -94.306801}
	case "XNN":
		return Location{"Xining", "Asia/Shanghai", 36.527500, 102.042999}
	case "XPL":
		return Location{"Comayagua", "America/Tegucigalpa", 14.382400, -87.621201}
	case "XQP":
		return Location{"Quepos", "America/Costa_Rica", 9.443160, -84.129799}
	case "XRY":
		return Location{"Jerez de la Forntera", "Europe/Madrid", 36.744598, -6.060110}
	case "XSC":
		return Location{"", "America/Grand_Turk", 21.515699, -71.528503}
	case "XSP":
		return Location{"Seletar", "Asia/Kuala_Lumpur", 1.416950, 103.867996}
	case "XTG":
		return Location{"", "Australia/Brisbane", -27.986401, 143.811005}
	case "XUZ":
		return Location{"Xuzhou", "Asia/Shanghai", 34.059056, 117.555278}
	case "YAA":
		return Location{"Anahim Lake", "America/Vancouver", 52.452499, -125.303001}
	case "YAB":
		return Location{"", "America/Rankin_Inlet", 73.005767, -85.042505}
	case "YAC":
		return Location{"Cat Lake", "America/Rainy_River", 51.727200, -91.824402}
	case "YAG":
		return Location{"Fort Frances", "America/Rainy_River", 48.654202, -93.439697}
	case "YAK":
		return Location{"Yakutat", "America/Yakutat", 59.503300, -139.660004}
	case "YAM":
		return Location{"Sault Ste Marie", "America/Detroit", 46.485001, -84.509399}
	case "YAP":
		return Location{"Yap Island", "Pacific/Chuuk", 9.498910, 138.082993}
	case "YAT":
		return Location{"Attawapiskat", "America/Nipigon", 52.927502, -82.431900}
	case "YAY":
		return Location{"St. Anthony", "America/St_Johns", 51.391899, -56.083099}
	case "YAZ":
		return Location{"Tofino", "America/Vancouver", 49.079826, -125.775604}
	case "YBB":
		return Location{"Kugaaruk", "America/Cambridge_Bay", 68.534401, -89.808098}
	case "YBE":
		return Location{"Uranium City", "America/Regina", 59.561401, -108.481003}
	case "YBG":
		return Location{"Bagotville", "America/Toronto", 48.330601, -70.996399}
	case "YBK":
		return Location{"Baker Lake", "America/Rankin_Inlet", 64.298897, -96.077797}
	case "YBL":
		return Location{"Campbell River", "America/Vancouver", 49.950802, -125.271004}
	case "YBP":
		return Location{"Yibin", "Asia/Shanghai", 28.800556, 104.545000}
	case "YBR":
		return Location{"Brandon", "America/Winnipeg", 49.910000, -99.951897}
	case "YBX":
		return Location{"Lourdes-De-Blanc-Sablon", "America/Blanc-Sablon", 51.443600, -57.185299}
	case "YCB":
		return Location{"Cambridge Bay", "America/Cambridge_Bay", 69.108101, -105.138000}
	case "YCD":
		return Location{"Nanaimo", "America/Vancouver", 49.052299, -123.870003}
	case "YCG":
		return Location{"Castlegar", "America/Vancouver", 49.296398, -117.632004}
	case "YCK":
		return Location{"Colville Lake", "America/Inuvik", 67.033302, -126.083000}
	case "YCO":
		return Location{"Kugluktuk", "America/Cambridge_Bay", 67.816704, -115.143997}
	case "YCS":
		return Location{"Chesterfield Inlet", "America/Rankin_Inlet", 63.346901, -90.731102}
	case "YCU":
		return Location{"Yuncheng", "Asia/Shanghai", 35.116391, 111.031389}
	case "YCY":
		return Location{"Clyde River", "America/Iqaluit", 70.486099, -68.516701}
	case "YDA":
		return Location{"Dawson City", "America/Dawson", 64.043098, -139.128006}
	case "YDF":
		return Location{"Deer Lake", "America/St_Johns", 49.210800, -57.391399}
	case "YDP":
		return Location{"Nain", "America/Goose_Bay", 56.549198, -61.680302}
	case "YEG":
		return Location{"Edmonton", "America/Edmonton", 53.309700, -113.580002}
	case "YEI":
		return Location{"Bursa", "Europe/Istanbul", 40.255199, 29.562599}
	case "YEK":
		return Location{"Arviat", "America/Rankin_Inlet", 61.094200, -94.070801}
	case "YEV":
		return Location{"Inuvik", "America/Inuvik", 68.304199, -133.483002}
	case "YFA":
		return Location{"Fort Albany", "America/Nipigon", 52.201401, -81.696899}
	case "YFB":
		return Location{"Iqaluit", "America/Iqaluit", 63.756401, -68.555801}
	case "YFC":
		return Location{"Fredericton", "America/Moncton", 45.868900, -66.537201}
	case "YFH":
		return Location{"Fort Hope", "America/Nipigon", 51.561901, -87.907799}
	case "YFJ":
		return Location{"Wekweeti", "America/Yellowknife", 64.190804, -114.077003}
	case "YFO":
		return Location{"Flin Flon", "America/Winnipeg", 54.678101, -101.681999}
	case "YFS":
		return Location{"Fort Simpson", "America/Inuvik", 61.760201, -121.237000}
	case "YFX":
		return Location{"St. Lewis", "America/St_Johns", 52.372799, -55.673901}
	case "YGH":
		return Location{"Fort Good Hope", "America/Inuvik", 66.240799, -128.651001}
	case "YGJ":
		return Location{"Yonago", "Asia/Tokyo", 35.492199, 133.235992}
	case "YGK":
		return Location{"Kingston", "America/Toronto", 44.225300, -76.596901}
	case "YGL":
		return Location{"La Grande Riviere", "America/Toronto", 53.625301, -77.704201}
	case "YGP":
		return Location{"Gaspe", "America/Toronto", 48.775299, -64.478600}
	case "YGR":
		return Location{"Iles-de-la-Madeleine", "America/Halifax", 47.424702, -61.778099}
	case "YGT":
		return Location{"Igloolik", "America/Iqaluit", 69.364700, -81.816101}
	case "YGW":
		return Location{"Kuujjuarapik", "America/Iqaluit", 55.281898, -77.765297}
	case "YGX":
		return Location{"Gillam", "America/Winnipeg", 56.357498, -94.710602}
	case "YGZ":
		return Location{"Grise Fiord", "America/Iqaluit", 76.426102, -82.909203}
	case "YHA":
		return Location{"Port Hope Simpson", "America/St_Johns", 52.528099, -56.286098}
	case "YHD":
		return Location{"Dryden", "America/Rainy_River", 49.831699, -92.744202}
	case "YHI":
		return Location{"Ulukhaktok", "America/Yellowknife", 70.762802, -117.806000}
	case "YHK":
		return Location{"Gjoa Haven", "America/Cambridge_Bay", 68.635597, -95.849701}
	case "YHM":
		return Location{"Hamilton", "America/Toronto", 43.173599, -79.934998}
	case "YHO":
		return Location{"Hopedale", "America/Goose_Bay", 55.448299, -60.228600}
	case "YHP":
		return Location{"Poplar Hill", "America/Rainy_River", 52.113300, -94.255600}
	case "YHR":
		return Location{"Chevery", "America/Blanc-Sablon", 50.468899, -59.636700}
	case "YHU":
		return Location{"Montreal", "America/Toronto", 45.517502, -73.416901}
	case "YHY":
		return Location{"Hay River", "America/Yellowknife", 60.839699, -115.782997}
	case "YHZ":
		return Location{"Halifax", "America/Halifax", 44.880798, -63.508598}
	case "YIA":
		return Location{"Yogyakarta", "Asia/Jakarta", -7.907459, 110.054480}
	case "YIF":
		return Location{"St-Augustin", "America/Blanc-Sablon", 51.211700, -58.658298}
	case "YIH":
		return Location{"Yichang", "Asia/Shanghai", 30.556550, 111.479988}
	case "YIK":
		return Location{"Ivujivik", "America/Iqaluit", 62.417301, -77.925301}
	case "YIN":
		return Location{"Yining", "Asia/Shanghai", 43.955799, 81.330299}
	case "YIO":
		return Location{"Pond Inlet", "America/Iqaluit", 72.683296, -77.966698}
	case "YIW":
		return Location{"Yiwu", "Asia/Shanghai", 29.344700, 120.031998}
	case "YKA":
		return Location{"Kamloops", "America/Vancouver", 50.702202, -120.444000}
	case "YKF":
		return Location{"Kitchener", "America/Toronto", 43.460800, -80.378601}
	case "YKG":
		return Location{"Kangirsuk", "America/Toronto", 60.027199, -69.999199}
	case "YKL":
		return Location{"Schefferville", "America/Toronto", 54.805302, -66.805298}
	case "YKM":
		return Location{"Yakima", "America/Los_Angeles", 46.568199, -120.543999}
	case "YKO":
		return Location{"Yuksekova", "Europe/Istanbul", 37.551667, 44.233611}
	case "YKQ":
		return Location{"Waskaganish", "America/Toronto", 51.473301, -78.758301}
	case "YKS":
		return Location{"Yakutsk", "Asia/Yakutsk", 62.093300, 129.770996}
	case "YKU":
		return Location{"Chisasibi", "America/Toronto", 53.805599, -78.916901}
	case "YLC":
		return Location{"Kimmirut", "America/Iqaluit", 62.849998, -69.883301}
	case "YLE":
		return Location{"Whati", "America/Yellowknife", 63.131699, -117.246002}
	case "YLH":
		return Location{"Lansdowne House", "America/Nipigon", 52.195599, -87.934196}
	case "YLL":
		return Location{"Lloydminster", "America/Edmonton", 53.309200, -110.072998}
	case "YLW":
		return Location{"Kelowna", "America/Vancouver", 49.956100, -119.377998}
	case "YMH":
		return Location{"Mary's Harbour", "America/St_Johns", 52.302799, -55.847198}
	case "YMM":
		return Location{"Fort McMurray", "America/Edmonton", 56.653301, -111.222000}
	case "YMN":
		return Location{"Makkovik", "America/Goose_Bay", 55.076900, -59.186401}
	case "YMO":
		return Location{"Moosonee", "America/Nipigon", 51.291100, -80.607803}
	case "YMT":
		return Location{"Chibougamau", "America/Toronto", 49.771900, -74.528099}
	case "YNA":
		return Location{"Natashquan", "America/Toronto", 50.189999, -61.789200}
	case "YNB":
		return Location{"", "Asia/Riyadh", 24.144199, 38.063400}
	case "YNC":
		return Location{"Wemindji", "America/Toronto", 53.010601, -78.831100}
	case "YNJ":
		return Location{"Yanji", "Asia/Shanghai", 42.882801, 129.451004}
	case "YNL":
		return Location{"Points North Landing", "America/Regina", 58.276699, -104.082001}
	case "YNO":
		return Location{"North Spirit Lake", "America/Rainy_River", 52.490002, -92.971100}
	case "YNS":
		return Location{"Nemiscau", "America/Toronto", 51.691101, -76.135597}
	case "YNT":
		return Location{"Yantai", "Asia/Shanghai", 37.401699, 121.372002}
	case "YNY":
		return Location{"Sokcho / Gangneung", "Asia/Seoul", 38.061298, 128.669006}
	case "YNZ":
		return Location{"Yancheng", "Asia/Shanghai", 33.425833, 120.203056}
	case "YOC":
		return Location{"Old Crow", "America/Dawson", 67.570602, -139.839005}
	case "YOG":
		return Location{"Ogoki Post", "America/Nipigon", 51.658600, -85.901703}
	case "YOJ":
		return Location{"High Level", "America/Edmonton", 58.621399, -117.165001}
	case "YOL":
		return Location{"Yola", "Africa/Lagos", 9.257550, 12.430400}
	case "YOW":
		return Location{"Ottawa", "America/Toronto", 45.322498, -75.669197}
	case "YPA":
		return Location{"Prince Albert", "America/Regina", 53.214199, -105.672997}
	case "YPH":
		return Location{"Inukjuak", "America/Toronto", 58.471901, -78.076897}
	case "YPJ":
		return Location{"Aupaluk", "America/Toronto", 59.296700, -69.599701}
	case "YPM":
		return Location{"Pikangikum", "America/Rainy_River", 51.819698, -93.973297}
	case "YPO":
		return Location{"Peawanuck", "America/Nipigon", 54.988098, -85.443298}
	case "YPR":
		return Location{"Prince Rupert", "America/Vancouver", 54.286098, -130.445007}
	case "YPW":
		return Location{"Powell River", "America/Vancouver", 49.834202, -124.500000}
	case "YPX":
		return Location{"Puvirnituq", "America/Toronto", 60.050598, -77.286903}
	case "YPY":
		return Location{"Fort Chipewyan", "America/Edmonton", 58.767200, -111.116997}
	case "YQB":
		return Location{"Quebec", "America/Toronto", 46.791100, -71.393303}
	case "YQC":
		return Location{"Quaqtaq", "America/Iqaluit", 61.046398, -69.617798}
	case "YQD":
		return Location{"The Pas", "America/Winnipeg", 53.971401, -101.091003}
	case "YQG":
		return Location{"Windsor", "America/Toronto", 42.275600, -82.955597}
	case "YQK":
		return Location{"Kenora", "America/Rainy_River", 49.788300, -94.363098}
	case "YQL":
		return Location{"Lethbridge", "America/Edmonton", 49.630299, -112.800003}
	case "YQM":
		return Location{"Moncton", "America/Moncton", 46.112202, -64.678596}
	case "YQQ":
		return Location{"Comox", "America/Vancouver", 49.710800, -124.887001}
	case "YQR":
		return Location{"Regina", "America/Regina", 50.431900, -104.666000}
	case "YQT":
		return Location{"Thunder Bay", "America/Thunder_Bay", 48.371899, -89.323898}
	case "YQU":
		return Location{"Grande Prairie", "America/Edmonton", 55.179699, -118.885002}
	case "YQX":
		return Location{"Gander", "America/St_Johns", 48.936901, -54.568100}
	case "YQY":
		return Location{"Sydney", "America/Glace_Bay", 46.161400, -60.047798}
	case "YQZ":
		return Location{"Quesnel", "America/Vancouver", 53.026100, -122.510002}
	case "YRA":
		return Location{"Gameti", "America/Yellowknife", 64.116096, -117.309998}
	case "YRB":
		return Location{"Resolute Bay", "America/Resolute", 74.716904, -94.969398}
	case "YRF":
		return Location{"Cartwright", "America/Goose_Bay", 53.682800, -57.041901}
	case "YRG":
		return Location{"Rigolet", "America/Goose_Bay", 54.179699, -58.457500}
	case "YRL":
		return Location{"Red Lake", "America/Rainy_River", 51.066898, -93.793098}
	case "YRT":
		return Location{"Rankin Inlet", "America/Rankin_Inlet", 62.811401, -92.115799}
	case "YSB":
		return Location{"Sudbury", "America/Toronto", 46.625000, -80.798897}
	case "YSF":
		return Location{"Stony Rapids", "America/Regina", 59.250301, -105.841003}
	case "YSG":
		return Location{"Lutselk'e", "America/Yellowknife", 62.418301, -110.681999}
	case "YSJ":
		return Location{"Saint John", "America/Moncton", 45.316101, -65.890297}
	case "YSK":
		return Location{"Sanikiluaq", "America/Iqaluit", 56.537800, -79.246696}
	case "YSM":
		return Location{"Fort Smith", "America/Edmonton", 60.020302, -111.961998}
	case "YTE":
		return Location{"Cape Dorset", "America/Iqaluit", 64.230003, -76.526703}
	case "YTH":
		return Location{"Thompson", "America/Winnipeg", 55.801102, -97.864197}
	case "YTQ":
		return Location{"Tasiujaq", "America/Toronto", 58.667801, -69.955803}
	case "YTS":
		return Location{"Timmins", "America/Toronto", 48.569698, -81.376701}
	case "YTZ":
		return Location{"Toronto", "America/Toronto", 43.627499, -79.396202}
	case "YUD":
		return Location{"Umiujaq", "America/Iqaluit", 56.536098, -76.518303}
	case "YUL":
		return Location{"Montreal", "America/Toronto", 45.470600, -73.740799}
	case "YUM":
		return Location{"Yuma", "America/Phoenix", 32.656601, -114.606003}
	case "YUS":
		return Location{"Yushu", "Asia/Shanghai", 32.836389, 97.036389}
	case "YUT":
		return Location{"Repulse Bay", "America/Rankin_Inlet", 66.521400, -86.224701}
	case "YUX":
		return Location{"Hall Beach", "America/Iqaluit", 68.776100, -81.243599}
	case "YUY":
		return Location{"Rouyn-Noranda", "America/Toronto", 48.206100, -78.835602}
	case "YVB":
		return Location{"Bonaventure", "America/Toronto", 48.071098, -65.460297}
	case "YVC":
		return Location{"La Ronge", "America/Regina", 55.151402, -105.262001}
	case "YVM":
		return Location{"Qikiqtarjuaq", "America/Pangnirtung", 67.545799, -64.031403}
	case "YVO":
		return Location{"Val-d'Or", "America/Toronto", 48.053299, -77.782799}
	case "YVP":
		return Location{"Kuujjuaq", "America/Toronto", 58.096100, -68.426903}
	case "YVQ":
		return Location{"Norman Wells", "America/Inuvik", 65.281601, -126.797997}
	case "YVR":
		return Location{"Vancouver", "America/Vancouver", 49.193901, -123.183998}
	case "YVZ":
		return Location{"Deer Lake", "America/Rainy_River", 52.655800, -94.061401}
	case "YWB":
		return Location{"Kangiqsujuaq", "America/Toronto", 61.588600, -71.929398}
	case "YWG":
		return Location{"Winnipeg", "America/Winnipeg", 49.910000, -97.239899}
	case "YWJ":
		return Location{"Deline", "America/Inuvik", 65.211098, -123.435997}
	case "YWK":
		return Location{"Wabush", "America/Goose_Bay", 52.921902, -66.864403}
	case "YWL":
		return Location{"Williams Lake", "America/Vancouver", 52.183102, -122.054001}
	case "YWP":
		return Location{"Webequie", "America/Nipigon", 52.959393, -87.374868}
	case "YXC":
		return Location{"Cranbrook", "America/Edmonton", 49.610802, -115.781998}
	case "YXE":
		return Location{"Saskatoon", "America/Regina", 52.170799, -106.699997}
	case "YXH":
		return Location{"Medicine Hat", "America/Edmonton", 50.018902, -110.721001}
	case "YXJ":
		return Location{"Fort St.John", "America/Dawson_Creek", 56.238098, -120.739998}
	case "YXL":
		return Location{"Sioux Lookout", "America/Rainy_River", 50.113899, -91.905296}
	case "YXN":
		return Location{"Whale Cove", "America/Rankin_Inlet", 62.240002, -92.598099}
	case "YXP":
		return Location{"Pangnirtung", "America/Pangnirtung", 66.144997, -65.713600}
	case "YXS":
		return Location{"Prince George", "America/Vancouver", 53.889400, -122.679001}
	case "YXT":
		return Location{"Terrace", "America/Vancouver", 54.468498, -128.576004}
	case "YXU":
		return Location{"London", "America/Toronto", 43.035599, -81.153900}
	case "YXX":
		return Location{"Abbotsford", "America/Los_Angeles", 49.025299, -122.361000}
	case "YXY":
		return Location{"Whitehorse", "America/Whitehorse", 60.709599, -135.067001}
	case "YYB":
		return Location{"North Bay", "America/Toronto", 46.363602, -79.422798}
	case "YYC":
		return Location{"Calgary", "America/Edmonton", 51.113899, -114.019997}
	case "YYD":
		return Location{"Smithers", "America/Vancouver", 54.824699, -127.182999}
	case "YYE":
		return Location{"Fort Nelson", "America/Fort_Nelson", 58.836399, -122.597000}
	case "YYF":
		return Location{"Penticton", "America/Vancouver", 49.463100, -119.601997}
	case "YYG":
		return Location{"Charlottetown", "America/Halifax", 46.290001, -63.121101}
	case "YYH":
		return Location{"Taloyoak", "America/Cambridge_Bay", 69.546700, -93.576698}
	case "YYJ":
		return Location{"Victoria", "America/Vancouver", 48.646900, -123.426003}
	case "YYQ":
		return Location{"Churchill", "America/Winnipeg", 58.739201, -94.065002}
	case "YYR":
		return Location{"Goose Bay", "America/Goose_Bay", 53.319199, -60.425800}
	case "YYT":
		return Location{"St. John's", "America/St_Johns", 47.618599, -52.751900}
	case "YYY":
		return Location{"Mont-Joli", "America/Toronto", 48.608601, -68.208099}
	case "YYZ":
		return Location{"Toronto", "America/Toronto", 43.677200, -79.630600}
	case "YZF":
		return Location{"Yellowknife", "America/Yellowknife", 62.462799, -114.440002}
	case "YZG":
		return Location{"Salluit", "America/Toronto", 62.179401, -75.667198}
	case "YZP":
		return Location{"Sandspit", "America/Vancouver", 53.254299, -131.813995}
	case "YZS":
		return Location{"Coral Harbour", "America/Atikokan", 64.193298, -83.359398}
	case "YZT":
		return Location{"Port Hardy", "America/Vancouver", 50.680599, -127.366997}
	case "YZV":
		return Location{"Sept-Iles", "America/Toronto", 50.223301, -66.265602}
	case "YZY":
		return Location{"Mackenzie", "America/Vancouver", 55.304401, -123.132004}
	case "YZZ":
		return Location{"Trail", "America/Vancouver", 49.055599, -117.609001}
	case "ZAD":
		return Location{"Zadar", "Europe/Zagreb", 44.108299, 15.346700}
	case "ZAG":
		return Location{"Zagreb", "Europe/Zagreb", 45.742901, 16.068800}
	case "ZAH":
		return Location{"Zahedan", "Asia/Tehran", 29.475700, 60.906200}
	case "ZAL":
		return Location{"Valdivia", "America/Santiago", -39.650002, -73.086098}
	case "ZAM":
		return Location{"Zamboanga City", "Asia/Manila", 6.922420, 122.059998}
	case "ZAT":
		return Location{"Zhaotong", "Asia/Shanghai", 27.325600, 103.754997}
	case "ZAZ":
		return Location{"Zaragoza", "Europe/Madrid", 41.666199, -1.041550}
	case "ZBF":
		return Location{"Bathurst", "America/Moncton", 47.629700, -65.738899}
	case "ZBR":
		return Location{"Chabahar", "Asia/Tehran", 25.443300, 60.382099}
	case "ZCL":
		return Location{"Zacatecas", "America/Mexico_City", 22.897100, -102.686996}
	case "ZCO":
		return Location{"Temuco", "America/Santiago", -38.766800, -72.637100}
	case "ZEL":
		return Location{"Bella Bella", "America/Vancouver", 52.185001, -128.156998}
	case "ZEM":
		return Location{"Eastmain River", "America/Toronto", 52.226398, -78.522499}
	case "ZER":
		return Location{"", "Asia/Kolkata", 27.588301, 93.828102}
	case "ZFD":
		return Location{"Fond-Du-Lac", "America/Regina", 59.334400, -107.181999}
	case "ZFN":
		return Location{"Tulita", "America/Inuvik", 64.909698, -125.572998}
	case "ZGU":
		return Location{"Gaua Island", "Pacific/Efate", -14.218100, 167.587006}
	case "ZHA":
		return Location{"Zhanjiang", "Asia/Shanghai", 21.214399, 110.358002}
	case "ZHY":
		return Location{"Zhongwei", "Asia/Shanghai", 37.572778, 105.154444}
	case "ZIA":
		return Location{"Trento", "Europe/Rome", 46.021400, 11.124200}
	case "ZIG":
		return Location{"Ziguinchor", "Africa/Dakar", 12.555600, -16.281799}
	case "ZIH":
		return Location{"Ixtapa", "America/Mexico_City", 17.601601, -101.460999}
	case "ZIX":
		return Location{"Zhigansk", "Asia/Yakutsk", 66.796501, 123.361000}
	case "ZKE":
		return Location{"Kashechewan", "America/Nipigon", 52.282501, -81.677803}
	case "ZKP":
		return Location{"Kasompe", "Africa/Lusaka", 65.736702, 150.705002}
	case "ZLO":
		return Location{"Manzanillo", "America/Mexico_City", 19.144800, -104.558998}
	case "ZLT":
		return Location{"La Tabatiere", "America/Blanc-Sablon", 50.830799, -58.975601}
	case "ZMT":
		return Location{"Masset", "America/Vancouver", 54.027500, -132.125000}
	case "ZNE":
		return Location{"Newman", "Australia/Perth", -23.417801, 119.803001}
	case "ZNZ":
		return Location{"Kiembi Samaki", "Africa/Dar_es_Salaam", -6.222020, 39.224899}
	case "ZOS":
		return Location{"Osorno", "America/Santiago", -40.611198, -73.060997}
	case "ZPB":
		return Location{"Sachigo Lake", "America/Rainy_River", 53.891102, -92.196404}
	case "ZQN":
		return Location{"Queenstown", "Pacific/Auckland", -45.021099, 168.738998}
	case "ZRH":
		return Location{"Zurich", "Europe/Zurich", 47.464699, 8.549170}
	case "ZSA":
		return Location{"San Salvador", "America/Nassau", 24.063299, -74.524002}
	case "ZSE":
		return Location{"St Pierre", "Indian/Reunion", -21.320900, 55.424999}
	case "ZSJ":
		return Location{"Sandy Lake", "America/Rainy_River", 53.064201, -93.344398}
	case "ZTA":
		return Location{"", "Pacific/Tahiti", -20.789700, -138.570007}
	case "ZTB":
		return Location{"Tete-a-la-Baleine", "America/Blanc-Sablon", 50.674400, -59.383598}
	case "ZTH":
		return Location{"Zakynthos Island", "Europe/Athens", 37.750900, 20.884300}
	case "ZUH":
		return Location{"Zhuhai", "Asia/Shanghai", 22.006399, 113.375999}
	case "ZUM":
		return Location{"Churchill Falls", "America/Goose_Bay", 53.561901, -64.106400}
	case "ZVK":
		return Location{"", "Asia/Bangkok", 16.556601, 104.760002}
	case "ZWL":
		return Location{"Wollaston Lake", "America/Regina", 58.106899, -103.171997}
	case "ZYI":
		return Location{"Zunyi", "Asia/Shanghai", 27.589500, 107.000700}
	case "ZYL":
		return Location{"Sylhet", "Asia/Dhaka", 24.963200, 91.866798}
	}
	return Location{"Not supported IATA Code", "Not supported IATA Code", 0, 0}
}
