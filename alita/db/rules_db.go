package db

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/utils/cache"
	"github.com/eko/gocache/lib/v4/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Rules struct {
	ChatId   int64  `bson:"_id,omitempty" json:"_id,omitempty"`
	Rules    string `bson:"rules" json:"rules" default:""`
	Private  bool   `bson:"privrules" json:"privrules"`
	RulesBtn string `bson:"rules_button,omitempty" json:"rules_button,omitempty"`
}

// check chat Flood Settings, used to get data before performing any operation
func checkRulesSetting(chatID int64) (rulesrc *Rules) {
	// Try cache first
	if cached, err := cache.Marshal.Get(cache.Context, chatID, new(Rules)); err == nil && cached != nil {
		return cached.(*Rules)
	}
	defRulesSrc := &Rules{ChatId: chatID, Rules: "", Private: false}
	errS := findOne(rulesColl, bson.M{"_id": chatID}).Decode(&rulesrc)
	if errS == mongo.ErrNoDocuments {
		rulesrc = defRulesSrc
		err := updateOne(rulesColl, bson.M{"_id": chatID}, rulesrc)
		if err != nil {
			log.Errorf("[Database] checkRulesSetting: %v - %d", err, chatID)
		}
	} else if errS != nil {
		rulesrc = defRulesSrc
		log.Errorf("[Database] checkRulesSetting: %v - %d", errS, chatID)
	}
	// Cache the result
	if rulesrc != nil {
		_ = cache.Marshal.Set(cache.Context, chatID, rulesrc, store.WithExpiration(10*time.Minute))
	}
	return rulesrc
}

func GetChatRulesInfo(chatId int64) *Rules {
	return checkRulesSetting(chatId)
}

func SetChatRules(chatId int64, rules string) {
	rulesUpdate := checkRulesSetting(chatId)
	rulesUpdate.Rules = rules
	err := updateOne(rulesColl, bson.M{"_id": chatId}, rulesUpdate)
	if err != nil {
		log.Errorf("[Database] SetChatRules: %v - %d", err, chatId)
	}
	// Update cache
	_ = cache.Marshal.Set(cache.Context, chatId, rulesUpdate, store.WithExpiration(10*time.Minute))
}

func SetChatRulesButton(chatId int64, rulesButton string) {
	rulesUpdate := checkRulesSetting(chatId)
	rulesUpdate.RulesBtn = rulesButton
	err := updateOne(rulesColl, bson.M{"_id": chatId}, rulesUpdate)
	if err != nil {
		log.Errorf("[Database] SetChatRulesButton: %v - %d", err, chatId)
	}
	// Update cache
	_ = cache.Marshal.Set(cache.Context, chatId, rulesUpdate, store.WithExpiration(10*time.Minute))
}

func SetPrivateRules(chatId int64, pref bool) {
	rulesUpdate := checkRulesSetting(chatId)
	rulesUpdate.Private = pref
	err := updateOne(rulesColl, bson.M{"_id": chatId}, rulesUpdate)
	if err != nil {
		log.Errorf("[Database] SetPrivateRules: %v - %d", err, chatId)
	}
	// Update cache
	_ = cache.Marshal.Set(cache.Context, chatId, rulesUpdate, store.WithExpiration(10*time.Minute))
}

func LoadRulesStats() (setRules, pvtRules int64) {
	setRules, clErr := countDocs(
		rulesColl,
		bson.M{
			"rules": bson.M{
				"$ne": "",
			},
		},
	)
	if clErr != nil {
		log.Errorf("[Database] LoadRulesStats: %v", clErr)
	}
	pvtRules, alErr := countDocs(
		rulesColl,
		bson.M{
			"privrules": true,
		},
	)
	if alErr != nil {
		log.Errorf("[Database] LoadRulesStats: %v", clErr)
	}
	return
}
