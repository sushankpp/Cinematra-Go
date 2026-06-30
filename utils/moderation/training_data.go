package moderation

type LabeledComment struct {
	Text  string
	Label string
}

var TrainingData = []LabeledComment{
	{
		Text:  "finally someone explained it properly",
		Label: "Safe",
	},
	{
		Text:  "nobody likes your videos",
		Label: "Problematic",
	},
	{
		Text:  "video unnecessarily long lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "same kura dherai repeat bhayo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "editing was clean and smooth",
		Label: "Safe",
	},
	{
		Text:  "haha that joke caught me off guard 😂",
		Label: "Safe",
	},
	{
		Text:  "this deserves more views",
		Label: "Safe",
	},
	{
		Text:  "camera focus issue thiyo ki?",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "loan chahiyo bina document contact me",
		Label: "Problematic",
	},
	{
		Text:  "could have been shorter",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "i made $1000 in one day ask me how",
		Label: "Problematic",
	},
	{
		Text:  "finally someone explained it properly",
		Label: "Safe",
	},
	{
		Text:  "i disagree with some points here",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "voice clear ra explanation ni ramro",
		Label: "Safe",
	},
	{
		Text:  "why did you skip that part?",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "background music ni ramro cha",
		Label: "Safe",
	},
	{
		Text:  "this actually helped me understand the topic",
		Label: "Safe",
	},
	{
		Text:  "bro yo kura chai milena jasto lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "this felt a bit biased",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "yo information double check garnu parcha hola",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "explanation ali confusing lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "telegram ma message gara easy income ko lagi",
		Label: "Problematic",
	},
	{
		Text:  "bro yo kura chai milena jasto lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "bro yo kura chai milena jasto lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "earn 5000 rupees daily from home dm me",
		Label: "Problematic",
	},
	{
		Text:  "bro yo kura chai milena jasto lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "camera focus issue thiyo ki?",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "telegram ma message gara easy income ko lagi",
		Label: "Problematic",
	},
	{
		Text:  "100% profit guaranteed no risk",
		Label: "Problematic",
	},
	{
		Text:  "yo information double check garnu parcha hola",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "bro yo kura chai milena jasto lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "i made $1000 in one day ask me how",
		Label: "Problematic",
	},
	{
		Text:  "free crypto giveaway click my profile",
		Label: "Problematic",
	},
	{
		Text:  "hot girls waiting click here",
		Label: "Problematic",
	},
	{
		Text:  "thanks for making this video",
		Label: "Safe",
	},
	{
		Text:  "yo part chai ekdam useful lagyo",
		Label: "Safe",
	},
	{
		Text:  "work from home and earn lakhs",
		Label: "Problematic",
	},
	{
		Text:  "respect for the effort put into this",
		Label: "Safe",
	},
	{
		Text:  "why did you skip that part?",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "worst creator ever delete your channel",
		Label: "Problematic",
	},
	{
		Text:  "join my whatsapp group for signals",
		Label: "Problematic",
	},
	{
		Text:  "haha that joke caught me off guard 😂",
		Label: "Safe",
	},
	{
		Text:  "thanks for making this video",
		Label: "Safe",
	},
	{
		Text:  "editing was clean and smooth",
		Label: "Safe",
	},
	{
		Text:  "aaja naya kura sikna paye",
		Label: "Safe",
	},
	{
		Text:  "i disagree with some points here",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "nobody likes your videos",
		Label: "Problematic",
	},
	{
		Text:  "this felt a bit biased",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "investment secret chahiyo bhane inbox gara",
		Label: "Problematic",
	},
	{
		Text:  "great work team",
		Label: "Safe",
	},
	{
		Text:  "work from home and earn lakhs",
		Label: "Problematic",
	},
	{
		Text:  "i made $1000 in one day ask me how",
		Label: "Problematic",
	},
	{
		Text:  "camera focus issue thiyo ki?",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "next part kaile aauxa?",
		Label: "Safe",
	},
	{
		Text:  "editing ali distracting lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "free crypto giveaway click my profile",
		Label: "Problematic",
	},
	{
		Text:  "click link in bio for free money",
		Label: "Problematic",
	},
	{
		Text:  "click link in bio for free money",
		Label: "Problematic",
	},
	{
		Text:  "great work team",
		Label: "Safe",
	},
	{
		Text:  "click link in bio for free money",
		Label: "Problematic",
	},
	{
		Text:  "join my whatsapp group for signals",
		Label: "Problematic",
	},
	{
		Text:  "source haru mention gareko bhaye ramro hunthyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "go disappear from youtube",
		Label: "Problematic",
	},
	{
		Text:  "this actually helped me understand the topic",
		Label: "Safe",
	},
	{
		Text:  "camera focus issue thiyo ki?",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "video unnecessarily long lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "next part kaile aauxa?",
		Label: "Safe",
	},
	{
		Text:  "thumbnail ali misleading jasto lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "click link in bio for free money",
		Label: "Problematic",
	},
	{
		Text:  "hot girls waiting click here",
		Label: "Problematic",
	},
	{
		Text:  "watched till the end, worth it",
		Label: "Safe",
	},
	{
		Text:  "this actually helped me understand the topic",
		Label: "Safe",
	},
	{
		Text:  "haha that joke caught me off guard 😂",
		Label: "Safe",
	},
	{
		Text:  "finally someone explained it properly",
		Label: "Safe",
	},
	{
		Text:  "next part kaile aauxa?",
		Label: "Safe",
	},
	{
		Text:  "respect for the effort put into this",
		Label: "Safe",
	},
	{
		Text:  "video unnecessarily long lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "same kura dherai repeat bhayo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "haha that joke caught me off guard 😂",
		Label: "Safe",
	},
	{
		Text:  "comment section ta war zone bhayexa 😅",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "audio ali low thiyo tara content ramro thiyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "thumbnail ali misleading jasto lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "telegram ma message gara easy income ko lagi",
		Label: "Problematic",
	},
	{
		Text:  "work from home and earn lakhs",
		Label: "Problematic",
	},
	{
		Text:  "video ekdam ramro xa dai 🔥",
		Label: "Safe",
	},
	{
		Text:  "this deserves more views",
		Label: "Safe",
	},
	{
		Text:  "could have been shorter",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "video ekdam ramro xa dai 🔥",
		Label: "Safe",
	},
	{
		Text:  "next part kaile aauxa?",
		Label: "Safe",
	},
	{
		Text:  "voice clear ra explanation ni ramro",
		Label: "Safe",
	},
	{
		Text:  "aaja naya kura sikna paye",
		Label: "Safe",
	},
	{
		Text:  "free crypto giveaway click my profile",
		Label: "Problematic",
	},
	{
		Text:  "visit my profile for betting tips",
		Label: "Problematic",
	},
	{
		Text:  "expected more details on this topic",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "aaja naya kura sikna paye",
		Label: "Safe",
	},
	{
		Text:  "audio ali low thiyo tara content ramro thiyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "work from home and earn lakhs",
		Label: "Problematic",
	},
	{
		Text:  "subscribe my channel and i will subscribe back",
		Label: "Problematic",
	},
	{
		Text:  "voice clear ra explanation ni ramro",
		Label: "Safe",
	},
	{
		Text:  "editing was clean and smooth",
		Label: "Safe",
	},
	{
		Text:  "visit my profile for betting tips",
		Label: "Problematic",
	},
	{
		Text:  "i made $1000 in one day ask me how",
		Label: "Problematic",
	},
	{
		Text:  "telegram ma message gara easy income ko lagi",
		Label: "Problematic",
	},
	{
		Text:  "why did you skip that part?",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "nobody likes your videos",
		Label: "Problematic",
	},
	{
		Text:  "aaja naya kura sikna paye",
		Label: "Safe",
	},
	{
		Text:  "telegram ma message gara easy income ko lagi",
		Label: "Problematic",
	},
	{
		Text:  "i disagree with some points here",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "not sure if this is correct",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "camera focus issue thiyo ki?",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "join my whatsapp group for signals",
		Label: "Problematic",
	},
	{
		Text:  "yo part chai ekdam useful lagyo",
		Label: "Safe",
	},
	{
		Text:  "i made $1000 in one day ask me how",
		Label: "Problematic",
	},
	{
		Text:  "join my whatsapp group for signals",
		Label: "Problematic",
	},
	{
		Text:  "nobody likes your videos",
		Label: "Problematic",
	},
	{
		Text:  "loan chahiyo bina document contact me",
		Label: "Problematic",
	},
	{
		Text:  "investment secret chahiyo bhane inbox gara",
		Label: "Problematic",
	},
	{
		Text:  "finally someone explained it properly",
		Label: "Safe",
	},
	{
		Text:  "this actually helped me understand the topic",
		Label: "Safe",
	},
	{
		Text:  "why did you skip that part?",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "pace ali fast bhayo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "editing was clean and smooth",
		Label: "Safe",
	},
	{
		Text:  "not sure if this is correct",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "respect for the effort put into this",
		Label: "Safe",
	},
	{
		Text:  "worst creator ever delete your channel",
		Label: "Problematic",
	},
	{
		Text:  "yo information double check garnu parcha hola",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "keep uploading videos like this bro",
		Label: "Safe",
	},
	{
		Text:  "explanation ali confusing lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "visit my profile for betting tips",
		Label: "Problematic",
	},
	{
		Text:  "camera focus issue thiyo ki?",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "editing ali distracting lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "i made $1000 in one day ask me how",
		Label: "Problematic",
	},
	{
		Text:  "you're an idiot",
		Label: "Problematic",
	},
	{
		Text:  "finally someone explained it properly",
		Label: "Safe",
	},
	{
		Text:  "telegram ma message gara easy income ko lagi",
		Label: "Problematic",
	},
	{
		Text:  "subscribed after watching this",
		Label: "Safe",
	},
	{
		Text:  "bro yo kura chai milena jasto lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "bro yo kura chai milena jasto lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "go disappear from youtube",
		Label: "Problematic",
	},
	{
		Text:  "haha that joke caught me off guard 😂",
		Label: "Safe",
	},
	{
		Text:  "subscribe my channel and i will subscribe back",
		Label: "Problematic",
	},
	{
		Text:  "nobody likes your videos",
		Label: "Problematic",
	},
	{
		Text:  "not sure if this is correct",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "expected more details on this topic",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "bhai timro content underrated xa",
		Label: "Safe",
	},
	{
		Text:  "comment section ta war zone bhayexa 😅",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "background music ni ramro cha",
		Label: "Safe",
	},
	{
		Text:  "editing ali distracting lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "watched till the end, worth it",
		Label: "Safe",
	},
	{
		Text:  "editing was clean and smooth",
		Label: "Safe",
	},
	{
		Text:  "editing was clean and smooth",
		Label: "Safe",
	},
	{
		Text:  "subscribe my channel and i will subscribe back",
		Label: "Problematic",
	},
	{
		Text:  "100% profit guaranteed no risk",
		Label: "Problematic",
	},
	{
		Text:  "same kura dherai repeat bhayo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "source haru mention gareko bhaye ramro hunthyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "explanation ali confusing lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "quality content as always",
		Label: "Safe",
	},
	{
		Text:  "this actually helped me understand the topic",
		Label: "Safe",
	},
	{
		Text:  "editing ali distracting lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "great work team",
		Label: "Safe",
	},
	{
		Text:  "voice clear ra explanation ni ramro",
		Label: "Safe",
	},
	{
		Text:  "watched till the end, worth it",
		Label: "Safe",
	},
	{
		Text:  "free crypto giveaway click my profile",
		Label: "Problematic",
	},
	{
		Text:  "check my channel for free followers",
		Label: "Problematic",
	},
	{
		Text:  "bhai timro content underrated xa",
		Label: "Safe",
	},
	{
		Text:  "audio ali low thiyo tara content ramro thiyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "thanks for making this video",
		Label: "Safe",
	},
	{
		Text:  "investment secret chahiyo bhane inbox gara",
		Label: "Problematic",
	},
	{
		Text:  "bhai timro content underrated xa",
		Label: "Safe",
	},
	{
		Text:  "thumbnail ali misleading jasto lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "check my channel for free followers",
		Label: "Problematic",
	},
	{
		Text:  "telegram ma message gara easy income ko lagi",
		Label: "Problematic",
	},
	{
		Text:  "click link in bio for free money",
		Label: "Problematic",
	},
	{
		Text:  "win free iphone now limited offer",
		Label: "Problematic",
	},
	{
		Text:  "go disappear from youtube",
		Label: "Problematic",
	},
	{
		Text:  "loan chahiyo bina document contact me",
		Label: "Problematic",
	},
	{
		Text:  "video ekdam ramro xa dai 🔥",
		Label: "Safe",
	},
	{
		Text:  "explanation ali confusing lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "i disagree with some points here",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "free recharge claim now",
		Label: "Problematic",
	},
	{
		Text:  "audio ali low thiyo tara content ramro thiyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "comment section ta war zone bhayexa 😅",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "great work team",
		Label: "Safe",
	},
	{
		Text:  "keep uploading videos like this bro",
		Label: "Safe",
	},
	{
		Text:  "respect for the effort put into this",
		Label: "Safe",
	},
	{
		Text:  "i disagree with some points here",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "same kura dherai repeat bhayo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "next part kaile aauxa?",
		Label: "Safe",
	},
	{
		Text:  "background music ni ramro cha",
		Label: "Safe",
	},
	{
		Text:  "editing was clean and smooth",
		Label: "Safe",
	},
	{
		Text:  "source haru mention gareko bhaye ramro hunthyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "this deserves more views",
		Label: "Safe",
	},
	{
		Text:  "bhai timro content underrated xa",
		Label: "Safe",
	},
	{
		Text:  "investment secret chahiyo bhane inbox gara",
		Label: "Problematic",
	},
	{
		Text:  "free recharge claim now",
		Label: "Problematic",
	},
	{
		Text:  "why did you skip that part?",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "earn 5000 rupees daily from home dm me",
		Label: "Problematic",
	},
	{
		Text:  "source haru mention gareko bhaye ramro hunthyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "i disagree with some points here",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "not sure if this is correct",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "subscribe my channel and i will subscribe back",
		Label: "Problematic",
	},
	{
		Text:  "respect for the effort put into this",
		Label: "Safe",
	},
	{
		Text:  "this felt a bit biased",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "same kura dherai repeat bhayo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "camera focus issue thiyo ki?",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "work from home and earn lakhs",
		Label: "Problematic",
	},
	{
		Text:  "quality paila ko video jasto lagena",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "keep uploading videos like this bro",
		Label: "Safe",
	},
	{
		Text:  "voice clear ra explanation ni ramro",
		Label: "Safe",
	},
	{
		Text:  "title herda arkai expect gareko thiye",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "why did you skip that part?",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "video unnecessarily long lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "worst creator ever delete your channel",
		Label: "Problematic",
	},
	{
		Text:  "voice clear ra explanation ni ramro",
		Label: "Safe",
	},
	{
		Text:  "subscribe my channel and i will subscribe back",
		Label: "Problematic",
	},
	{
		Text:  "subscribe my channel and i will subscribe back",
		Label: "Problematic",
	},
	{
		Text:  "free recharge claim now",
		Label: "Problematic",
	},
	{
		Text:  "quality paila ko video jasto lagena",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "next part kaile aauxa?",
		Label: "Safe",
	},
	{
		Text:  "thanks for making this video",
		Label: "Safe",
	},
	{
		Text:  "explanation ali confusing lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "win free iphone now limited offer",
		Label: "Problematic",
	},
	{
		Text:  "camera focus issue thiyo ki?",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "voice clear ra explanation ni ramro",
		Label: "Safe",
	},
	{
		Text:  "win free iphone now limited offer",
		Label: "Problematic",
	},
	{
		Text:  "background music ni ramro cha",
		Label: "Safe",
	},
	{
		Text:  "investment secret chahiyo bhane inbox gara",
		Label: "Problematic",
	},
	{
		Text:  "this deserves more views",
		Label: "Safe",
	},
	{
		Text:  "thumbnail ali misleading jasto lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "video ekdam ramro xa dai 🔥",
		Label: "Safe",
	},
	{
		Text:  "yo part chai ekdam useful lagyo",
		Label: "Safe",
	},
	{
		Text:  "editing ali distracting lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "respect for the effort put into this",
		Label: "Safe",
	},
	{
		Text:  "title herda arkai expect gareko thiye",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "subscribed after watching this",
		Label: "Safe",
	},
	{
		Text:  "bahut accha video tha, maza aa gaya",
		Label: "Safe",
	},
	{
		Text:  "nobody likes your videos",
		Label: "Problematic",
	},
	{
		Text:  "investment secret chahiyo bhane inbox gara",
		Label: "Problematic",
	},
	{
		Text:  "could have been shorter",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "expected more details on this topic",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "yo information double check garnu parcha hola",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "earn 5000 rupees daily from home dm me",
		Label: "Problematic",
	},
	{
		Text:  "loan chahiyo bina document contact me",
		Label: "Problematic",
	},
	{
		Text:  "expected more details on this topic",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "hot girls waiting click here",
		Label: "Problematic",
	},
	{
		Text:  "audio ali low thiyo tara content ramro thiyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "i disagree with some points here",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "same kura dherai repeat bhayo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "hot girls waiting click here",
		Label: "Problematic",
	},
	{
		Text:  "voice clear ra explanation ni ramro",
		Label: "Safe",
	},
	{
		Text:  "telegram ma message gara easy income ko lagi",
		Label: "Problematic",
	},
	{
		Text:  "bahut accha video tha, maza aa gaya",
		Label: "Safe",
	},
	{
		Text:  "click link in bio for free money",
		Label: "Problematic",
	},
	{
		Text:  "editing was clean and smooth",
		Label: "Safe",
	},
	{
		Text:  "explanation ali confusing lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "video unnecessarily long lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "subscribed after watching this",
		Label: "Safe",
	},
	{
		Text:  "check my channel for free followers",
		Label: "Problematic",
	},
	{
		Text:  "i made $1000 in one day ask me how",
		Label: "Problematic",
	},
	{
		Text:  "video unnecessarily long lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "go disappear from youtube",
		Label: "Problematic",
	},
	{
		Text:  "background music ni ramro cha",
		Label: "Safe",
	},
	{
		Text:  "you're an idiot",
		Label: "Problematic",
	},
	{
		Text:  "worst creator ever delete your channel",
		Label: "Problematic",
	},
	{
		Text:  "not sure if this is correct",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "you're an idiot",
		Label: "Problematic",
	},
	{
		Text:  "could have been shorter",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "respect for the effort put into this",
		Label: "Safe",
	},
	{
		Text:  "click link in bio for free money",
		Label: "Problematic",
	},
	{
		Text:  "title herda arkai expect gareko thiye",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "watched till the end, worth it",
		Label: "Safe",
	},
	{
		Text:  "why did you skip that part?",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "you're an idiot",
		Label: "Problematic",
	},
	{
		Text:  "aaja naya kura sikna paye",
		Label: "Safe",
	},
	{
		Text:  "free crypto giveaway click my profile",
		Label: "Problematic",
	},
	{
		Text:  "hot girls waiting click here",
		Label: "Problematic",
	},
	{
		Text:  "could have been shorter",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "bhai timro content underrated xa",
		Label: "Safe",
	},
	{
		Text:  "background music ni ramro cha",
		Label: "Safe",
	},
	{
		Text:  "comment section ta war zone bhayexa 😅",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "worst creator ever delete your channel",
		Label: "Problematic",
	},
	{
		Text:  "video unnecessarily long lagyo",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "loan chahiyo bina document contact me",
		Label: "Problematic",
	},
	{
		Text:  "thanks for making this video",
		Label: "Safe",
	},
	{
		Text:  "work from home and earn lakhs",
		Label: "Problematic",
	},
	{
		Text:  "comment section ta war zone bhayexa 😅",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "bhai timro content underrated xa",
		Label: "Safe",
	},
	{
		Text:  "voice clear ra explanation ni ramro",
		Label: "Safe",
	},
	{
		Text:  "bahut accha video tha, maza aa gaya",
		Label: "Safe",
	},
	{
		Text:  "keep uploading videos like this bro",
		Label: "Safe",
	},
	{
		Text:  "comment section ta war zone bhayexa 😅",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "i disagree with some points here",
		Label: "PotentiallyProblematic",
	},
	{
		Text:  "investment secret chahiyo bhane inbox gara",
		Label: "Problematic",
	},
	{
		Text:  "nobody likes your videos",
		Label: "Problematic",
	},
	{
		Text:  "keep uploading videos like this bro",
		Label: "Safe",
	},
	{
		Text:  "check my channel for free followers",
		Label: "Problematic",
	},
	{
		Text:  "yo part chai ekdam useful lagyo",
		Label: "Safe",
	},
	{
		Text:  "this deserves more views",
		Label: "Safe",
	},
	{
		Text:  "nobody likes your videos",
		Label: "Problematic",
	},
	{
		Text:  "work from home and earn lakhs",
		Label: "Problematic",
	},
}
