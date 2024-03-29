package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

type Metamon struct {
	address    string
	privateKey string
	c          *resty.Client
	backoff    Backoff
}

type Backoff struct {
	minDelay time.Duration
	maxDelay time.Duration
}

func (b *Backoff) next(attempt int) time.Duration {
	if attempt < 0 {
		return b.minDelay
	}

	minf := float64(b.minDelay)
	durf := minf * math.Pow(1.5, float64(attempt))
	durf = durf + rand.Float64()*minf

	delay := time.Duration(durf)
	if delay > b.maxDelay {
		return b.maxDelay
	}

	return delay
}

func New(privateKey string) *Metamon {
	address, err := getPublicKey(privateKey)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("login wallet address:", address)
	return &Metamon{
		address:    address,
		privateKey: privateKey,
		c:          resty.New(),
		backoff:    Backoff{maxDelay: 3 * time.Second, minDelay: 1 * time.Second},
	}
}

func prefixHash(data []byte) common.Hash {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	return crypto.Keccak256Hash([]byte(msg))
}

func (m *Metamon) sign(msg string) (string, error) {
	privateKey, err := crypto.HexToECDSA(m.privateKey)
	if err != nil {
		return "", err
	}

	data := []byte(msg)
	hash := prefixHash(data)

	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return "", err
	}

	// https://stackoverflow.com/questions/69762108/implementing-ethereum-personal-sign-eip-191-from-go-ethereum-gives-different-s
	signature[64] += 27

	return hexutil.Encode(signature), nil
}

func (m *Metamon) Login(sign, msg string) error {
	params := map[string]string{
		"address": m.address,
		"sign":    sign,
		"msg":     msg,
		"network": "1",
	}

	resp, err := m.c.R().SetQueryParams(params).Post(loginURL)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return errors.New("response err")
	}

	var out LoginRes
	if err := decodeResponse(resp.Body(), &out); err != nil {
		return err
	}

	log.Println(out)

	m.setHeaders(out.AccessToken)

	return nil
}

func (m *Metamon) setHeaders(token string) {
	headers := map[string]string{
		"content-type": "multipart/form-data; boundary=----WebKitFormBoundaryBoisGGEqBQMlOG7a",
		"accesstoken":  token,
		"user_agent":   "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.99 Safari/537.36",
	}
	m.c = m.c.SetHeaders(headers)
}

func (m *Metamon) GetObjects(address, metamonId, battleLevel string) ([]*Monster, error) {
	params := map[string]string{
		"address":   address,
		"metamonId": metamonId,
		"front":     battleLevel,
	}

	resp, err := m.c.R().SetQueryParams(params).Post(getObjectURL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, errors.New("response err")
	}

	var out MonsterObject
	if err := decodeResponse(resp.Body(), &out); err != nil {
		return nil, err
	}

	return out.Objects, nil
}

func (m *Metamon) StartPay(from, to, battleLevel string) error {
	params := map[string]string{
		"address":     m.address,
		"battleLevel": battleLevel,
		"monsterA":    from,
		"monsterB":    to,
	}

	resp, err := m.c.R().SetQueryParams(params).Post(startPayURL)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return errors.New("response err")
	}

	var ret PayResult
	if err := decodeResponse(resp.Body(), &ret); err != nil {
		return err
	}

	if ret.Pay == true {
		log.Printf("pay %d raca success, please go to battle \n", ret.Amount)
	} else {
		log.Printf("pay failed! please try again \n")
		return errors.New("pay failed")
	}

	return nil
}

func (m *Metamon) StartBattle(from, to, battleLevel string) (*BattleResult, error) {
	params := map[string]string{
		"address":     m.address,
		"battleLevel": battleLevel,
		"monsterA":    from,
		"monsterB":    to,
	}

	resp, err := m.c.R().SetQueryParams(params).Post(startBattleURL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, errors.New("response err")
	}

	var ret BattleResult
	if err := decodeResponse(resp.Body(), &ret); err != nil {
		return nil, err
	}

	return &ret, nil
}

func (m *Metamon) getWalletProperty() ([]*Monster, error) {
	params := map[string]string{
		"address":   m.address,
		"orderType": "-1",
	}

	resp, err := m.c.R().SetQueryParams(params).Post(getWalletPropertyList)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, errors.New("response err")
	}

	var monsters MonsterList
	if err := decodeResponse(resp.Body(), &monsters); err != nil {
		return nil, err
	}

	return monsters.Monsters, nil
}

func decodeResponse(in []byte, out interface{}) error {
	var ret Response
	if err := json.Unmarshal(in, &ret); err != nil {
		return err
	}

	if ret.Code != "SUCCESS" {
		log.Println(ret.Message)
	}

	data, err := json.Marshal(ret.Data)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}

	return nil
}

func getPublicKey(pk string) (string, error) {
	bytes, err := hexutil.Decode(fmt.Sprintf("0x%s", pk))
	if err != nil {
		return "", err
	}

	privateKey, err := crypto.ToECDSA(bytes)
	if err != nil {
		return "", err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", errors.New("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	return address, nil
}

func getBattleLevel(level int64) string {
	if level < 21 {
		return "1"
	} else if level >= 21 && level < 41 {
		return "2"
	} else {
		return "3"
	}
}

func (m *Metamon) UpdateMonster(monsterId string) error {
	params := map[string]string{
		"nftId":   monsterId,
		"address": m.address,
	}

	resp, err := m.c.R().SetQueryParams(params).Post(updateMonsterURL)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return errors.New("response err")
	}

	var ret Response
	if err := json.Unmarshal(resp.Body(), &ret); err != nil {
		return err
	}

	log.Println("update monster: ", ret.Code)
	return nil
}

func (m *Metamon) GetBag() ([]*BagItem, error) {
	params := map[string]string{
		"address": m.address,
	}

	resp, err := m.c.R().SetQueryParams(params).Post(checkBagURL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, errors.New("response err")
	}

	var ret Bag
	if err := decodeResponse(resp.Body(), &ret); err != nil {
		return nil, err
	}

	return ret.Item, nil
}

func (m *Metamon) mint() error {
	params := map[string]string{
		"address": m.address,
	}

	resp, err := m.c.R().SetQueryParams(params).Post(composeMonsterEggURL)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return errors.New("response err")
	}

	var ret Response
	if err := json.Unmarshal(resp.Body(), &ret); err != nil {
		return err
	}

	log.Println("mint egg result: ", ret.Code)
	return nil
}

func (m *Metamon) Mint() error {
	bagItems, err := m.GetBag()
	if err != nil {
		return err
	}

	for _, item := range bagItems {
		if item.Type != 1 {
			continue
		}

		num, _ := strconv.Atoi(item.Number)
		if num < 1000 {
			return nil
		}

		if err := m.mint(); err != nil {
			return err
		}
	}

	return nil
}

func (m *Metamon) ResetMonsterEXP(monsterId string) error {
	params := map[string]string{
		"address": m.address,
		"nftId":   monsterId,
	}

	resp, err := m.c.R().SetQueryParams(params).Post(resetMonstEXPURL)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return errors.New("response err")
	}

	var ret Response
	if err := json.Unmarshal(resp.Body(), &ret); err != nil {
		return err
	}

	log.Println("monster reset EXP result: ", ret.Code)
	return nil
}

func (m *Metamon) GetTeamList() ([]*Team, error) {
	params := map[string]string{
		"address":    m.address,
		"page":       "1",
		"pageSize":   "20",
		"orderField": "monsterNum",
	}

	resp, err := m.c.R().SetQueryParams(params).Post(teamListURL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, errors.New("response err")
	}

	var ret TeamListResponse
	if err := decodeResponse(resp.Body(), &ret); err != nil {
		return nil, err
	}

	return ret.TeamList, nil
}

func (m *Metamon) GetScreenMetamon(teamId, scaThreshold string) (*ScreenMetamon, error) {
	params := map[string]string{
		"address":      m.address,
		"scaThreshold": scaThreshold,
		"teamId":       teamId,
		"minSca":       "-1",
		"nftId":        "-1",
		"pageSize":     "50",
	}

	resp, err := m.c.R().SetQueryParams(params).Post(screenMetamonURL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, errors.New("response err")
	}

	var ret ScreenMetamon
	if err := decodeResponse(resp.Body(), &ret); err != nil {
		return nil, err
	}

	return &ret, nil
}

func (m *Metamon) JoinTeam(teamId string, ids []*nftId) error {
	req := &JoinRequest{
		Address:  m.address,
		Metamons: ids,
		TeamId:   teamId,
	}

	resp, err := m.c.R().SetBody(req).Post(joinTeamUrl)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return errors.New("response err")
	}

	var ret Response
	if err := json.Unmarshal(resp.Body(), &ret); err != nil {
		return err
	}

	log.Println("Join team result: ", ret.Code)
	return nil
}

func (m *Metamon) tryToJoinTeam() error {
	teamList, err := m.GetTeamList()
	if err != nil {
		return err
	}

	bagItems, err := m.GetBag()
	if err != nil {
		return err
	}

	for _, item := range bagItems {
		if item.Type != 6 {
			continue
		}
		bpNum, _ := strconv.ParseInt(item.Number, 10, 64)
		if bpNum > 0 {
			for _, team := range teamList {
				if team.LockTeam {
					continue
				}

				if team.MonsterScaThreshold > 305 {
					continue
				}

				threshold := strconv.FormatInt(team.MonsterScaThreshold, 10)
				metamon, err := m.GetScreenMetamon(team.Id, threshold)
				if err != nil {
					return err
				}

				var nftIds []*nftId
				for _, item := range metamon.Monsters {
					nftIds = append(nftIds, &nftId{
						nftId: item.Id,
					})
				}

				if err := m.JoinTeam(team.Id, nftIds); err != nil {
					return err
				}

				break
			}
		}
	}

	return nil
}

func main() {
	msg := fmt.Sprintf("LogIn-%s", uuid.New())
	m := New(os.Getenv("WALLET_PRIVATE_KEY"))

	sign, err := m.sign(msg)
	if err != nil {
		log.Fatalln("sign: ", err)
		return
	}

	if err := m.Login(sign, msg); err != nil {
		log.Fatalf("login: %v", err)
	}

	myMonsters, err := m.getWalletProperty()
	if err != nil {
		log.Fatalln("get monster failed: ", err)
		return
	}

	// 元兽王国
	m.tryToJoinTeam()

	for _, monster := range myMonsters {
		var winCount int64

		exp := monster.Exp

		for i := 0; i < int(monster.Tear); i++ {
			if exp >= monster.ExpMax {
				if monster.Level >= 60 {
					fmt.Println("monster need to reset EXP")
					err := m.ResetMonsterEXP(monster.Id)
					if err != nil {
						log.Fatalln(err)
					}
				} else {
					if err := m.UpdateMonster(monster.Id); err != nil {
						log.Println(err)
					} else {
						exp = 0
					}
				}
			}

			battleLevel := getBattleLevel(monster.Level)
			monsters, err := m.GetObjects(monster.Owner, monster.Id, battleLevel)
			if err != nil {
				log.Fatalln(err)
			}

			sort.Slice(monsters, func(i, j int) bool {
				return monsters[i].Sca < monsters[j].Sca
			})

			best := monsters[0]
			log.Printf("The weakest monster: Id: %s, Rarity:%s, Sca: %d, LV: %s", best.TokenId, best.Rarity, best.Sca, battleLevel)

			time.Sleep(m.backoff.next(i))

			if err := m.StartPay(monster.Id, best.Id, battleLevel); err != nil {
				log.Fatalln("start pay failed: ", err)
			}

			ret, err := m.StartBattle(monster.Id, best.Id, battleLevel)
			if err != nil {
				log.Fatalln("start battle failed:", err)
			}

			if ret.ChallengeResult {
				winCount++
			}

			exp += ret.ChallengeExp

			log.Printf("battle result: isWin: %t fragmentNum: %d  EXP: %d \n", ret.ChallengeResult, ret.BpFragmentNum, ret.ChallengeExp)
		}

		if monster.Tear > 0 {
			log.Printf("total battles: %d, win: %d, winRate:%d%% \n", monster.Tear, winCount, winCount*100/monster.Tear)
		}
	}

	if err := m.Mint(); err != nil {
		log.Printf("mint egg: %v\n", err)
	}
}
