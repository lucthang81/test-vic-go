package l

import (
	//	"fmt"

	"github.com/vic/vic_go/zconfig"
)

const (
	/*
		for i in xrange(1, 1100):
		    print 'M{0:04d} = "M{0:04}"'.format(i)
	*/
	M0001 = "M0001"
	M0002 = "M0002"
	M0003 = "M0003"
	M0004 = "M0004"
	M0005 = "M0005"
	M0006 = "M0006"
	M0007 = "M0007"
	M0008 = "M0008"
	M0009 = "M0009"
	M0010 = "M0010"
	M0011 = "M0011"
	M0012 = "M0012"
	M0013 = "M0013"
	M0014 = "M0014"
	M0015 = "M0015"
	M0016 = "M0016"
	M0017 = "M0017"
	M0018 = "M0018"
	M0019 = "M0019"
	M0020 = "M0020"
	M0021 = "M0021"
	M0022 = "M0022"
	M0023 = "M0023"
	M0024 = "M0024"
	M0025 = "M0025"
	M0026 = "M0026"
	M0027 = "M0027"
	M0028 = "M0028"
	M0029 = "M0029"
	M0030 = "M0030"
	M0031 = "M0031"
	M0032 = "M0032"
	M0033 = "M0033"
	M0034 = "M0034"
	M0035 = "M0035"
	M0036 = "M0036"
	M0037 = "M0037"
	M0038 = "M0038"
	M0039 = "M0039"
	M0040 = "M0040"
	M0041 = "M0041"
	M0042 = "M0042"
	M0043 = "M0043"
	M0044 = "M0044"
	M0045 = "M0045"
	M0046 = "M0046"
	M0047 = "M0047"
	M0048 = "M0048"
	M0049 = "M0049"
	M0050 = "M0050"
	M0051 = "M0051"
	M0052 = "M0052"
	M0053 = "M0053"
	M0054 = "M0054"
	M0055 = "M0055"
	M0056 = "M0056"
	M0057 = "M0057"
	M0058 = "M0058"
	M0059 = "M0059"
	M0060 = "M0060"
	M0061 = "M0061"
	M0062 = "M0062"
	M0063 = "M0063"
	M0064 = "M0064"
	M0065 = "M0065"
	M0066 = "M0066"
	M0067 = "M0067"
	M0068 = "M0068"
	M0069 = "M0069"
	M0070 = "M0070"
	M0071 = "M0071"
	M0072 = "M0072"
	M0073 = "M0073"
	M0074 = "M0074"
	M0075 = "M0075"
	M0076 = "M0076"
	M0077 = "M0077"
	M0078 = "M0078"
	M0079 = "M0079"
	M0080 = "M0080"
	M0081 = "M0081"
	M0082 = "M0082"
	M0083 = "M0083"
	M0084 = "M0084"
	M0085 = "M0085"
	M0086 = "M0086"
	M0087 = "M0087"
	M0088 = "M0088"
	M0089 = "M0089"
	M0090 = "M0090"
	M0091 = "M0091"
	M0092 = "M0092"
	M0093 = "M0093"
	M0094 = "M0094"
	M0095 = "M0095"
	M0096 = "M0096"
	M0097 = "M0097"
	M0098 = "M0098"
	M0099 = "M0099"
	M0100 = "M0100"
	M0101 = "M0101"
	M0102 = "M0102"
	M0103 = "M0103"
	M0104 = "M0104"
	M0105 = "M0105"
	M0106 = "M0106"
	M0107 = "M0107"
	M0108 = "M0108"
	M0109 = "M0109"
	M0110 = "M0110"
	M0111 = "M0111"
	M0112 = "M0112"
	M0113 = "M0113"
	M0114 = "M0114"
	M0115 = "M0115"
	M0116 = "M0116"
	M0117 = "M0117"
	M0118 = "M0118"
	M0119 = "M0119"
	M0120 = "M0120"
	M0121 = "M0121"
	M0122 = "M0122"
	M0123 = "M0123"
	M0124 = "M0124"
	M0125 = "M0125"
	M0126 = "M0126"
	M0127 = "M0127"
	M0128 = "M0128"
	M0129 = "M0129"
	M0130 = "M0130"
	M0131 = "M0131"
	M0132 = "M0132"
	M0133 = "M0133"
	M0134 = "M0134"
	M0135 = "M0135"
	M0136 = "M0136"
	M0137 = "M0137"
	M0138 = "M0138"
	M0139 = "M0139"
	M0140 = "M0140"
	M0141 = "M0141"
	M0142 = "M0142"
	M0143 = "M0143"
	M0144 = "M0144"
	M0145 = "M0145"
	M0146 = "M0146"
	M0147 = "M0147"
	M0148 = "M0148"
	M0149 = "M0149"
	M0150 = "M0150"
	M0151 = "M0151"
	M0152 = "M0152"
	M0153 = "M0153"
	M0154 = "M0154"
	M0155 = "M0155"
	M0156 = "M0156"
	M0157 = "M0157"
	M0158 = "M0158"
	M0159 = "M0159"
	M0160 = "M0160"
	M0161 = "M0161"
	M0162 = "M0162"
	M0163 = "M0163"
	M0164 = "M0164"
	M0165 = "M0165"
	M0166 = "M0166"
	M0167 = "M0167"
	M0168 = "M0168"
	M0169 = "M0169"
	M0170 = "M0170"
	M0171 = "M0171"
	M0172 = "M0172"
	M0173 = "M0173"
	M0174 = "M0174"
	M0175 = "M0175"
	M0176 = "M0176"
	M0177 = "M0177"
	M0178 = "M0178"
	M0179 = "M0179"
	M0180 = "M0180"
	M0181 = "M0181"
	M0182 = "M0182"
	M0183 = "M0183"
	M0184 = "M0184"
	M0185 = "M0185"
	M0186 = "M0186"
	M0187 = "M0187"
	M0188 = "M0188"
	M0189 = "M0189"
	M0190 = "M0190"
	M0191 = "M0191"
	M0192 = "M0192"
	M0193 = "M0193"
	M0194 = "M0194"
	M0195 = "M0195"
	M0196 = "M0196"
	M0197 = "M0197"
	M0198 = "M0198"
	M0199 = "M0199"
	M0200 = "M0200"
	M0201 = "M0201"
	M0202 = "M0202"
	M0203 = "M0203"
	M0204 = "M0204"
	M0205 = "M0205"
	M0206 = "M0206"
	M0207 = "M0207"
	M0208 = "M0208"
	M0209 = "M0209"
	M0210 = "M0210"
	M0211 = "M0211"
	M0212 = "M0212"
	M0213 = "M0213"
	M0214 = "M0214"
	M0215 = "M0215"
	M0216 = "M0216"
	M0217 = "M0217"
	M0218 = "M0218"
	M0219 = "M0219"
	M0220 = "M0220"
	M0221 = "M0221"
	M0222 = "M0222"
	M0223 = "M0223"
	M0224 = "M0224"
	M0225 = "M0225"
	M0226 = "M0226"
	M0227 = "M0227"
	M0228 = "M0228"
	M0229 = "M0229"
	M0230 = "M0230"
	M0231 = "M0231"
	M0232 = "M0232"
	M0233 = "M0233"
	M0234 = "M0234"
	M0235 = "M0235"
	M0236 = "M0236"
	M0237 = "M0237"
	M0238 = "M0238"
	M0239 = "M0239"
	M0240 = "M0240"
	M0241 = "M0241"
	M0242 = "M0242"
	M0243 = "M0243"
	M0244 = "M0244"
	M0245 = "M0245"
	M0246 = "M0246"
	M0247 = "M0247"
	M0248 = "M0248"
	M0249 = "M0249"
	M0250 = "M0250"
	M0251 = "M0251"
	M0252 = "M0252"
	M0253 = "M0253"
	M0254 = "M0254"
	M0255 = "M0255"
	M0256 = "M0256"
	M0257 = "M0257"
	M0258 = "M0258"
	M0259 = "M0259"
	M0260 = "M0260"
	M0261 = "M0261"
	M0262 = "M0262"
	M0263 = "M0263"
	M0264 = "M0264"
	M0265 = "M0265"
	M0266 = "M0266"
	M0267 = "M0267"
	M0268 = "M0268"
	M0269 = "M0269"
	M0270 = "M0270"
	M0271 = "M0271"
	M0272 = "M0272"
	M0273 = "M0273"
	M0274 = "M0274"
	M0275 = "M0275"
	M0276 = "M0276"
	M0277 = "M0277"
	M0278 = "M0278"
	M0279 = "M0279"
	M0280 = "M0280"
	M0281 = "M0281"
	M0282 = "M0282"
	M0283 = "M0283"
	M0284 = "M0284"
	M0285 = "M0285"
	M0286 = "M0286"
	M0287 = "M0287"
	M0288 = "M0288"
	M0289 = "M0289"
	M0290 = "M0290"
	M0291 = "M0291"
	M0292 = "M0292"
	M0293 = "M0293"
	M0294 = "M0294"
	M0295 = "M0295"
	M0296 = "M0296"
	M0297 = "M0297"
	M0298 = "M0298"
	M0299 = "M0299"
	M0300 = "M0300"
	M0301 = "M0301"
	M0302 = "M0302"
	M0303 = "M0303"
	M0304 = "M0304"
	M0305 = "M0305"
	M0306 = "M0306"
	M0307 = "M0307"
	M0308 = "M0308"
	M0309 = "M0309"
	M0310 = "M0310"
	M0311 = "M0311"
	M0312 = "M0312"
	M0313 = "M0313"
	M0314 = "M0314"
	M0315 = "M0315"
	M0316 = "M0316"
	M0317 = "M0317"
	M0318 = "M0318"
	M0319 = "M0319"
	M0320 = "M0320"
	M0321 = "M0321"
	M0322 = "M0322"
	M0323 = "M0323"
	M0324 = "M0324"
	M0325 = "M0325"
	M0326 = "M0326"
	M0327 = "M0327"
	M0328 = "M0328"
	M0329 = "M0329"
	M0330 = "M0330"
	M0331 = "M0331"
	M0332 = "M0332"
	M0333 = "M0333"
	M0334 = "M0334"
	M0335 = "M0335"
	M0336 = "M0336"
	M0337 = "M0337"
	M0338 = "M0338"
	M0339 = "M0339"
	M0340 = "M0340"
	M0341 = "M0341"
	M0342 = "M0342"
	M0343 = "M0343"
	M0344 = "M0344"
	M0345 = "M0345"
	M0346 = "M0346"
	M0347 = "M0347"
	M0348 = "M0348"
	M0349 = "M0349"
	M0350 = "M0350"
	M0351 = "M0351"
	M0352 = "M0352"
	M0353 = "M0353"
	M0354 = "M0354"
	M0355 = "M0355"
	M0356 = "M0356"
	M0357 = "M0357"
	M0358 = "M0358"
	M0359 = "M0359"
	M0360 = "M0360"
	M0361 = "M0361"
	M0362 = "M0362"
	M0363 = "M0363"
	M0364 = "M0364"
	M0365 = "M0365"
	M0366 = "M0366"
	M0367 = "M0367"
	M0368 = "M0368"
	M0369 = "M0369"
	M0370 = "M0370"
	M0371 = "M0371"
	M0372 = "M0372"
	M0373 = "M0373"
	M0374 = "M0374"
	M0375 = "M0375"
	M0376 = "M0376"
	M0377 = "M0377"
	M0378 = "M0378"
	M0379 = "M0379"
	M0380 = "M0380"
	M0381 = "M0381"
	M0382 = "M0382"
	M0383 = "M0383"
	M0384 = "M0384"
	M0385 = "M0385"
	M0386 = "M0386"
	M0387 = "M0387"
	M0388 = "M0388"
	M0389 = "M0389"
	M0390 = "M0390"
	M0391 = "M0391"
	M0392 = "M0392"
	M0393 = "M0393"
	M0394 = "M0394"
	M0395 = "M0395"
	M0396 = "M0396"
	M0397 = "M0397"
	M0398 = "M0398"
	M0399 = "M0399"
	M0400 = "M0400"
	M0401 = "M0401"
	M0402 = "M0402"
	M0403 = "M0403"
	M0404 = "M0404"
	M0405 = "M0405"
	M0406 = "M0406"
	M0407 = "M0407"
	M0408 = "M0408"
	M0409 = "M0409"
	M0410 = "M0410"
	M0411 = "M0411"
	M0412 = "M0412"
	M0413 = "M0413"
	M0414 = "M0414"
	M0415 = "M0415"
	M0416 = "M0416"
	M0417 = "M0417"
	M0418 = "M0418"
	M0419 = "M0419"
	M0420 = "M0420"
	M0421 = "M0421"
	M0422 = "M0422"
	M0423 = "M0423"
	M0424 = "M0424"
	M0425 = "M0425"
	M0426 = "M0426"
	M0427 = "M0427"
	M0428 = "M0428"
	M0429 = "M0429"
	M0430 = "M0430"
	M0431 = "M0431"
	M0432 = "M0432"
	M0433 = "M0433"
	M0434 = "M0434"
	M0435 = "M0435"
	M0436 = "M0436"
	M0437 = "M0437"
	M0438 = "M0438"
	M0439 = "M0439"
	M0440 = "M0440"
	M0441 = "M0441"
	M0442 = "M0442"
	M0443 = "M0443"
	M0444 = "M0444"
	M0445 = "M0445"
	M0446 = "M0446"
	M0447 = "M0447"
	M0448 = "M0448"
	M0449 = "M0449"
	M0450 = "M0450"
	M0451 = "M0451"
	M0452 = "M0452"
	M0453 = "M0453"
	M0454 = "M0454"
	M0455 = "M0455"
	M0456 = "M0456"
	M0457 = "M0457"
	M0458 = "M0458"
	M0459 = "M0459"
	M0460 = "M0460"
	M0461 = "M0461"
	M0462 = "M0462"
	M0463 = "M0463"
	M0464 = "M0464"
	M0465 = "M0465"
	M0466 = "M0466"
	M0467 = "M0467"
	M0468 = "M0468"
	M0469 = "M0469"
	M0470 = "M0470"
	M0471 = "M0471"
	M0472 = "M0472"
	M0473 = "M0473"
	M0474 = "M0474"
	M0475 = "M0475"
	M0476 = "M0476"
	M0477 = "M0477"
	M0478 = "M0478"
	M0479 = "M0479"
	M0480 = "M0480"
	M0481 = "M0481"
	M0482 = "M0482"
	M0483 = "M0483"
	M0484 = "M0484"
	M0485 = "M0485"
	M0486 = "M0486"
	M0487 = "M0487"
	M0488 = "M0488"
	M0489 = "M0489"
	M0490 = "M0490"
	M0491 = "M0491"
	M0492 = "M0492"
	M0493 = "M0493"
	M0494 = "M0494"
	M0495 = "M0495"
	M0496 = "M0496"
	M0497 = "M0497"
	M0498 = "M0498"
	M0499 = "M0499"
	M0500 = "M0500"
	M0501 = "M0501"
	M0502 = "M0502"
	M0503 = "M0503"
	M0504 = "M0504"
	M0505 = "M0505"
	M0506 = "M0506"
	M0507 = "M0507"
	M0508 = "M0508"
	M0509 = "M0509"
	M0510 = "M0510"
	M0511 = "M0511"
	M0512 = "M0512"
	M0513 = "M0513"
	M0514 = "M0514"
	M0515 = "M0515"
	M0516 = "M0516"
	M0517 = "M0517"
	M0518 = "M0518"
	M0519 = "M0519"
	M0520 = "M0520"
	M0521 = "M0521"
	M0522 = "M0522"
	M0523 = "M0523"
	M0524 = "M0524"
	M0525 = "M0525"
	M0526 = "M0526"
	M0527 = "M0527"
	M0528 = "M0528"
	M0529 = "M0529"
	M0530 = "M0530"
	M0531 = "M0531"
	M0532 = "M0532"
	M0533 = "M0533"
	M0534 = "M0534"
	M0535 = "M0535"
	M0536 = "M0536"
	M0537 = "M0537"
	M0538 = "M0538"
	M0539 = "M0539"
	M0540 = "M0540"
	M0541 = "M0541"
	M0542 = "M0542"
	M0543 = "M0543"
	M0544 = "M0544"
	M0545 = "M0545"
	M0546 = "M0546"
	M0547 = "M0547"
	M0548 = "M0548"
	M0549 = "M0549"
	M0550 = "M0550"
	M0551 = "M0551"
	M0552 = "M0552"
	M0553 = "M0553"
	M0554 = "M0554"
	M0555 = "M0555"
	M0556 = "M0556"
	M0557 = "M0557"
	M0558 = "M0558"
	M0559 = "M0559"
	M0560 = "M0560"
	M0561 = "M0561"
	M0562 = "M0562"
	M0563 = "M0563"
	M0564 = "M0564"
	M0565 = "M0565"
	M0566 = "M0566"
	M0567 = "M0567"
	M0568 = "M0568"
	M0569 = "M0569"
	M0570 = "M0570"
	M0571 = "M0571"
	M0572 = "M0572"
	M0573 = "M0573"
	M0574 = "M0574"
	M0575 = "M0575"
	M0576 = "M0576"
	M0577 = "M0577"
	M0578 = "M0578"
	M0579 = "M0579"
	M0580 = "M0580"
	M0581 = "M0581"
	M0582 = "M0582"
	M0583 = "M0583"
	M0584 = "M0584"
	M0585 = "M0585"
	M0586 = "M0586"
	M0587 = "M0587"
	M0588 = "M0588"
	M0589 = "M0589"
	M0590 = "M0590"
	M0591 = "M0591"
	M0592 = "M0592"
	M0593 = "M0593"
	M0594 = "M0594"
	M0595 = "M0595"
	M0596 = "M0596"
	M0597 = "M0597"
	M0598 = "M0598"
	M0599 = "M0599"
	M0600 = "M0600"
	M0601 = "M0601"
	M0602 = "M0602"
	M0603 = "M0603"
	M0604 = "M0604"
	M0605 = "M0605"
	M0606 = "M0606"
	M0607 = "M0607"
	M0608 = "M0608"
	M0609 = "M0609"
	M0610 = "M0610"
	M0611 = "M0611"
	M0612 = "M0612"
	M0613 = "M0613"
	M0614 = "M0614"
	M0615 = "M0615"
	M0616 = "M0616"
	M0617 = "M0617"
	M0618 = "M0618"
	M0619 = "M0619"
	M0620 = "M0620"
	M0621 = "M0621"
	M0622 = "M0622"
	M0623 = "M0623"
	M0624 = "M0624"
	M0625 = "M0625"
	M0626 = "M0626"
	M0627 = "M0627"
	M0628 = "M0628"
	M0629 = "M0629"
	M0630 = "M0630"
	M0631 = "M0631"
	M0632 = "M0632"
	M0633 = "M0633"
	M0634 = "M0634"
	M0635 = "M0635"
	M0636 = "M0636"
	M0637 = "M0637"
	M0638 = "M0638"
	M0639 = "M0639"
	M0640 = "M0640"
	M0641 = "M0641"
	M0642 = "M0642"
	M0643 = "M0643"
	M0644 = "M0644"
	M0645 = "M0645"
	M0646 = "M0646"
	M0647 = "M0647"
	M0648 = "M0648"
	M0649 = "M0649"
	M0650 = "M0650"
	M0651 = "M0651"
	M0652 = "M0652"
	M0653 = "M0653"
	M0654 = "M0654"
	M0655 = "M0655"
	M0656 = "M0656"
	M0657 = "M0657"
	M0658 = "M0658"
	M0659 = "M0659"
	M0660 = "M0660"
	M0661 = "M0661"
	M0662 = "M0662"
	M0663 = "M0663"
	M0664 = "M0664"
	M0665 = "M0665"
	M0666 = "M0666"
	M0667 = "M0667"
	M0668 = "M0668"
	M0669 = "M0669"
	M0670 = "M0670"
	M0671 = "M0671"
	M0672 = "M0672"
	M0673 = "M0673"
	M0674 = "M0674"
	M0675 = "M0675"
	M0676 = "M0676"
	M0677 = "M0677"
	M0678 = "M0678"
	M0679 = "M0679"
	M0680 = "M0680"
	M0681 = "M0681"
	M0682 = "M0682"
	M0683 = "M0683"
	M0684 = "M0684"
	M0685 = "M0685"
	M0686 = "M0686"
	M0687 = "M0687"
	M0688 = "M0688"
	M0689 = "M0689"
	M0690 = "M0690"
	M0691 = "M0691"
	M0692 = "M0692"
	M0693 = "M0693"
	M0694 = "M0694"
	M0695 = "M0695"
	M0696 = "M0696"
	M0697 = "M0697"
	M0698 = "M0698"
	M0699 = "M0699"
	M0700 = "M0700"
	M0701 = "M0701"
	M0702 = "M0702"
	M0703 = "M0703"
	M0704 = "M0704"
	M0705 = "M0705"
	M0706 = "M0706"
	M0707 = "M0707"
	M0708 = "M0708"
	M0709 = "M0709"
	M0710 = "M0710"
	M0711 = "M0711"
	M0712 = "M0712"
	M0713 = "M0713"
	M0714 = "M0714"
	M0715 = "M0715"
	M0716 = "M0716"
	M0717 = "M0717"
	M0718 = "M0718"
	M0719 = "M0719"
	M0720 = "M0720"
	M0721 = "M0721"
	M0722 = "M0722"
	M0723 = "M0723"
	M0724 = "M0724"
	M0725 = "M0725"
	M0726 = "M0726"
	M0727 = "M0727"
	M0728 = "M0728"
	M0729 = "M0729"
	M0730 = "M0730"
	M0731 = "M0731"
	M0732 = "M0732"
	M0733 = "M0733"
	M0734 = "M0734"
	M0735 = "M0735"
	M0736 = "M0736"
	M0737 = "M0737"
	M0738 = "M0738"
	M0739 = "M0739"
	M0740 = "M0740"
	M0741 = "M0741"
	M0742 = "M0742"
	M0743 = "M0743"
	M0744 = "M0744"
	M0745 = "M0745"
	M0746 = "M0746"
	M0747 = "M0747"
	M0748 = "M0748"
	M0749 = "M0749"
	M0750 = "M0750"
	M0751 = "M0751"
	M0752 = "M0752"
	M0753 = "M0753"
	M0754 = "M0754"
	M0755 = "M0755"
	M0756 = "M0756"
	M0757 = "M0757"
	M0758 = "M0758"
	M0759 = "M0759"
	M0760 = "M0760"
	M0761 = "M0761"
	M0762 = "M0762"
	M0763 = "M0763"
	M0764 = "M0764"
	M0765 = "M0765"
	M0766 = "M0766"
	M0767 = "M0767"
	M0768 = "M0768"
	M0769 = "M0769"
	M0770 = "M0770"
	M0771 = "M0771"
	M0772 = "M0772"
	M0773 = "M0773"
	M0774 = "M0774"
	M0775 = "M0775"
	M0776 = "M0776"
	M0777 = "M0777"
	M0778 = "M0778"
	M0779 = "M0779"
	M0780 = "M0780"
	M0781 = "M0781"
	M0782 = "M0782"
	M0783 = "M0783"
	M0784 = "M0784"
	M0785 = "M0785"
	M0786 = "M0786"
	M0787 = "M0787"
	M0788 = "M0788"
	M0789 = "M0789"
	M0790 = "M0790"
	M0791 = "M0791"
	M0792 = "M0792"
	M0793 = "M0793"
	M0794 = "M0794"
	M0795 = "M0795"
	M0796 = "M0796"
	M0797 = "M0797"
	M0798 = "M0798"
	M0799 = "M0799"
	M0800 = "M0800"
	M0801 = "M0801"
	M0802 = "M0802"
	M0803 = "M0803"
	M0804 = "M0804"
	M0805 = "M0805"
	M0806 = "M0806"
	M0807 = "M0807"
	M0808 = "M0808"
	M0809 = "M0809"
	M0810 = "M0810"
	M0811 = "M0811"
	M0812 = "M0812"
	M0813 = "M0813"
	M0814 = "M0814"
	M0815 = "M0815"
	M0816 = "M0816"
	M0817 = "M0817"
	M0818 = "M0818"
	M0819 = "M0819"
	M0820 = "M0820"
	M0821 = "M0821"
	M0822 = "M0822"
	M0823 = "M0823"
	M0824 = "M0824"
	M0825 = "M0825"
	M0826 = "M0826"
	M0827 = "M0827"
	M0828 = "M0828"
	M0829 = "M0829"
	M0830 = "M0830"
	M0831 = "M0831"
	M0832 = "M0832"
	M0833 = "M0833"
	M0834 = "M0834"
	M0835 = "M0835"
	M0836 = "M0836"
	M0837 = "M0837"
	M0838 = "M0838"
	M0839 = "M0839"
	M0840 = "M0840"
	M0841 = "M0841"
	M0842 = "M0842"
	M0843 = "M0843"
	M0844 = "M0844"
	M0845 = "M0845"
	M0846 = "M0846"
	M0847 = "M0847"
	M0848 = "M0848"
	M0849 = "M0849"
	M0850 = "M0850"
	M0851 = "M0851"
	M0852 = "M0852"
	M0853 = "M0853"
	M0854 = "M0854"
	M0855 = "M0855"
	M0856 = "M0856"
	M0857 = "M0857"
	M0858 = "M0858"
	M0859 = "M0859"
	M0860 = "M0860"
	M0861 = "M0861"
	M0862 = "M0862"
	M0863 = "M0863"
	M0864 = "M0864"
	M0865 = "M0865"
	M0866 = "M0866"
	M0867 = "M0867"
	M0868 = "M0868"
	M0869 = "M0869"
	M0870 = "M0870"
	M0871 = "M0871"
	M0872 = "M0872"
	M0873 = "M0873"
	M0874 = "M0874"
	M0875 = "M0875"
	M0876 = "M0876"
	M0877 = "M0877"
	M0878 = "M0878"
	M0879 = "M0879"
	M0880 = "M0880"
	M0881 = "M0881"
	M0882 = "M0882"
	M0883 = "M0883"
	M0884 = "M0884"
	M0885 = "M0885"
	M0886 = "M0886"
	M0887 = "M0887"
	M0888 = "M0888"
	M0889 = "M0889"
	M0890 = "M0890"
	M0891 = "M0891"
	M0892 = "M0892"
	M0893 = "M0893"
	M0894 = "M0894"
	M0895 = "M0895"
	M0896 = "M0896"
	M0897 = "M0897"
	M0898 = "M0898"
	M0899 = "M0899"
	M0900 = "M0900"
	M0901 = "M0901"
	M0902 = "M0902"
	M0903 = "M0903"
	M0904 = "M0904"
	M0905 = "M0905"
	M0906 = "M0906"
	M0907 = "M0907"
	M0908 = "M0908"
	M0909 = "M0909"
	M0910 = "M0910"
	M0911 = "M0911"
	M0912 = "M0912"
	M0913 = "M0913"
	M0914 = "M0914"
	M0915 = "M0915"
	M0916 = "M0916"
	M0917 = "M0917"
	M0918 = "M0918"
	M0919 = "M0919"
	M0920 = "M0920"
	M0921 = "M0921"
	M0922 = "M0922"
	M0923 = "M0923"
	M0924 = "M0924"
	M0925 = "M0925"
	M0926 = "M0926"
	M0927 = "M0927"
	M0928 = "M0928"
	M0929 = "M0929"
	M0930 = "M0930"
	M0931 = "M0931"
	M0932 = "M0932"
	M0933 = "M0933"
	M0934 = "M0934"
	M0935 = "M0935"
	M0936 = "M0936"
	M0937 = "M0937"
	M0938 = "M0938"
	M0939 = "M0939"
	M0940 = "M0940"
	M0941 = "M0941"
	M0942 = "M0942"
	M0943 = "M0943"
	M0944 = "M0944"
	M0945 = "M0945"
	M0946 = "M0946"
	M0947 = "M0947"
	M0948 = "M0948"
	M0949 = "M0949"
	M0950 = "M0950"
	M0951 = "M0951"
	M0952 = "M0952"
	M0953 = "M0953"
	M0954 = "M0954"
	M0955 = "M0955"
	M0956 = "M0956"
	M0957 = "M0957"
	M0958 = "M0958"
	M0959 = "M0959"
	M0960 = "M0960"
	M0961 = "M0961"
	M0962 = "M0962"
	M0963 = "M0963"
	M0964 = "M0964"
	M0965 = "M0965"
	M0966 = "M0966"
	M0967 = "M0967"
	M0968 = "M0968"
	M0969 = "M0969"
	M0970 = "M0970"
	M0971 = "M0971"
	M0972 = "M0972"
	M0973 = "M0973"
	M0974 = "M0974"
	M0975 = "M0975"
	M0976 = "M0976"
	M0977 = "M0977"
	M0978 = "M0978"
	M0979 = "M0979"
	M0980 = "M0980"
	M0981 = "M0981"
	M0982 = "M0982"
	M0983 = "M0983"
	M0984 = "M0984"
	M0985 = "M0985"
	M0986 = "M0986"
	M0987 = "M0987"
	M0988 = "M0988"
	M0989 = "M0989"
	M0990 = "M0990"
	M0991 = "M0991"
	M0992 = "M0992"
	M0993 = "M0993"
	M0994 = "M0994"
	M0995 = "M0995"
	M0996 = "M0996"
	M0997 = "M0997"
	M0998 = "M0998"
	M0999 = "M0999"
	M1000 = "M1000"
	M1001 = "M1001"
	M1002 = "M1002"
	M1003 = "M1003"
	M1004 = "M1004"
	M1005 = "M1005"
	M1006 = "M1006"
	M1007 = "M1007"
	M1008 = "M1008"
	M1009 = "M1009"
	M1010 = "M1010"
	M1011 = "M1011"
	M1012 = "M1012"
	M1013 = "M1013"
	M1014 = "M1014"
	M1015 = "M1015"
	M1016 = "M1016"
	M1017 = "M1017"
	M1018 = "M1018"
	M1019 = "M1019"
	M1020 = "M1020"
	M1021 = "M1021"
	M1022 = "M1022"
	M1023 = "M1023"
	M1024 = "M1024"
	M1025 = "M1025"
	M1026 = "M1026"
	M1027 = "M1027"
	M1028 = "M1028"
	M1029 = "M1029"
	M1030 = "M1030"
	M1031 = "M1031"
	M1032 = "M1032"
	M1033 = "M1033"
	M1034 = "M1034"
	M1035 = "M1035"
	M1036 = "M1036"
	M1037 = "M1037"
	M1038 = "M1038"
	M1039 = "M1039"
	M1040 = "M1040"
	M1041 = "M1041"
	M1042 = "M1042"
	M1043 = "M1043"
	M1044 = "M1044"
	M1045 = "M1045"
	M1046 = "M1046"
	M1047 = "M1047"
	M1048 = "M1048"
	M1049 = "M1049"
	M1050 = "M1050"
	M1051 = "M1051"
	M1052 = "M1052"
	M1053 = "M1053"
	M1054 = "M1054"
	M1055 = "M1055"
	M1056 = "M1056"
	M1057 = "M1057"
	M1058 = "M1058"
	M1059 = "M1059"
	M1060 = "M1060"
	M1061 = "M1061"
	M1062 = "M1062"
	M1063 = "M1063"
	M1064 = "M1064"
	M1065 = "M1065"
	M1066 = "M1066"
	M1067 = "M1067"
	M1068 = "M1068"
	M1069 = "M1069"
	M1070 = "M1070"
	M1071 = "M1071"
	M1072 = "M1072"
	M1073 = "M1073"
	M1074 = "M1074"
	M1075 = "M1075"
	M1076 = "M1076"
	M1077 = "M1077"
	M1078 = "M1078"
	M1079 = "M1079"
	M1080 = "M1080"
	M1081 = "M1081"
	M1082 = "M1082"
	M1083 = "M1083"
	M1084 = "M1084"
	M1085 = "M1085"
	M1086 = "M1086"
	M1087 = "M1087"
	M1088 = "M1088"
	M1089 = "M1089"
	M1090 = "M1090"
	M1091 = "M1091"
	M1092 = "M1092"
	M1093 = "M1093"
	M1094 = "M1094"
	M1095 = "M1095"
	M1096 = "M1096"
	M1097 = "M1097"
	M1098 = "M1098"
	M1099 = "M1099"
)

// map msgName to msgContent
var mapMessages map[string]string

func init() {
	// fmt.Println("zconfig.Language", zconfig.Language)
	if zconfig.Language == zconfig.LANG_ENGLISH {
		mapMessages = mapMessagesEnglish
	} else if zconfig.Language == zconfig.LANG_VIETNAMESE {
		mapMessages = mapMessagesVietnamese
	} else {
		mapMessages = make(map[string]string)
	}
}

// get messageContent from messageName
func Get(msgName string) string {
	return mapMessages[msgName]
}
