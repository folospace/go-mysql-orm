package main

import (
    "github.com/folospace/go-mysql-orm/orm"
    "testing"
    "time"
)

var tdb, _ = orm.OpenMysql("rfamro@tcp(mysql-rfam-public.ebi.ac.uk:4497)/Rfam?parseTime=true&charset=utf8mb4&loc=Asia%2FShanghai")

var FamilyTable2 = orm.NewQuery(Family2{}, tdb)

type Family2 struct {
    RfamAcc            string    `json:"rfam_acc" orm:"rfam_acc,varchar(7),primary,unique"`
    RfamId             string    `json:"rfam_id" orm:"rfam_id,varchar(40),index"`
    AutoWiki           uint      `json:"auto_wiki" orm:"auto_wiki,int(10) unsigned,index"`
    Description        *string   `json:"description" orm:"description,varchar(75),null" default:"NULL"`
    Author             *string   `json:"author" orm:"author,tinytext,null"`
    SeedSource         *string   `json:"seed_source" orm:"seed_source,tinytext,null"`
    GatheringCutoff    *float64  `json:"gathering_cutoff" orm:"gathering_cutoff,double(5,2),null" default:"NULL"`
    TrustedCutoff      *float64  `json:"trusted_cutoff" orm:"trusted_cutoff,double(5,2),null" default:"NULL"`
    NoiseCutoff        *float64  `json:"noise_cutoff" orm:"noise_cutoff,double(5,2),null" default:"NULL"`
    Comment            *string   `json:"comment" orm:"comment,longtext,null"`
    PreviousId         *string   `json:"previous_id" orm:"previous_id,tinytext,null"`
    Cmbuild            *string   `json:"cmbuild" orm:"cmbuild,tinytext,null"`
    Cmcalibrate        *string   `json:"cmcalibrate" orm:"cmcalibrate,tinytext,null"`
    Cmsearch           *string   `json:"cmsearch" orm:"cmsearch,tinytext,null"`
    NumSeed            *int64    `json:"num_seed" orm:"num_seed,bigint(20),null" default:"NULL"`
    NumFull            *int64    `json:"num_full" orm:"num_full,bigint(20),null" default:"NULL"`
    NumGenomeSeq       *int64    `json:"num_genome_seq" orm:"num_genome_seq,bigint(20),null" default:"NULL"`
    NumRefseq          *int64    `json:"num_refseq" orm:"num_refseq,bigint(20),null" default:"NULL"`
    Type               *string   `json:"type" orm:"type,varchar(50),null" default:"NULL"`
    StructureSource    *string   `json:"structure_source" orm:"structure_source,tinytext,null"`
    NumberOfSpecies    *int64    `json:"number_of_species" orm:"number_of_species,bigint(20),null" default:"NULL"`
    Number3dStructures *int      `json:"number_3d_structures" orm:"number_3d_structures,int(11),null" default:"NULL"`
    NumPseudonokts     *int      `json:"num_pseudonokts" orm:"num_pseudonokts,int(11),null" default:"NULL"`
    TaxSeed            *string   `json:"tax_seed" orm:"tax_seed,mediumtext,null"`
    EcmliLambda        *float64  `json:"ecmli_lambda" orm:"ecmli_lambda,double(10,5),null" default:"NULL"`
    EcmliMu            *float64  `json:"ecmli_mu" orm:"ecmli_mu,double(10,5),null" default:"NULL"`
    EcmliCalDb         *int      `json:"ecmli_cal_db" orm:"ecmli_cal_db,mediumint(9),null" default:"0"`
    EcmliCalHits       *int      `json:"ecmli_cal_hits" orm:"ecmli_cal_hits,mediumint(9),null" default:"0"`
    Maxl               *int      `json:"maxl" orm:"maxl,mediumint(9),null" default:"0"`
    Clen               *int      `json:"clen" orm:"clen,mediumint(9),null" default:"0"`
    MatchPairNode      *int8     `json:"match_pair_node" orm:"match_pair_node,tinyint(1),null" default:"0"`
    HmmTau             *float64  `json:"hmm_tau" orm:"hmm_tau,double(10,5),null" default:"NULL"`
    HmmLambda          *float64  `json:"hmm_lambda" orm:"hmm_lambda,double(10,5),null" default:"NULL"`
    Created            time.Time `json:"created" orm:"created,datetime"`
    Updated            time.Time `json:"updated" orm:"updated,timestamp" default:"CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
}

func (Family2) TableName() string {
    return "family"
}

func (Family2) DatabaseName() string {
    return "Rfam"
}

func TestSelect(t *testing.T) {
    t.Run("mysql_version", func(t *testing.T) {
        var data string
        query := FamilyTable2.Raw("select version()").GetTo(&data)

        t.Log(data)
        t.Log(query.Sql())
        t.Log(query.Error())
    })
    t.Run("query_timeout", func(t *testing.T) {
        var data map[string]int
        query := FamilyTable2.Raw("show variables like '%timeout%'").GetTo(&data)

        t.Log(data)
        t.Log(query.Sql())
        t.Log(query.Error())
    })
    t.Run("query_table_sql", func(t *testing.T) {
        var data map[string]string
        query := FamilyTable2.Raw("show create table " + FamilyTable2.T.TableName()).GetTo(&data)

        t.Log(data)
        t.Log(query.Sql())
        t.Log(query.Error())
    })
    t.Run("count_total", func(t *testing.T) {
        var data int64
        query := FamilyTable2.Select("count(*)").GetTo(&data)
        t.Log(data)

        data, query = FamilyTable2.GetCount()
        t.Log(data)

        t.Log(query.Sql())
        t.Log(query.Error())
    })
    t.Run("count_distinct_total", func(t *testing.T) {
        var data int64
        query := FamilyTable2.Select("count(distinct(type))").GetTo(&data)
        t.Log(data)

        data, query = FamilyTable2.GroupBy(&FamilyTable2.T.Type).GetCount()
        t.Log(data)

        t.Log(query.Sql())
        t.Log(query.Error())
    })
    t.Run("result_to_map", func(t *testing.T) {
        var data map[string][]string
        query := FamilyTable2.Select(&FamilyTable2.T.Type, &FamilyTable2.T.RfamAcc).Limit(20).GetTo(&data)
        t.Log(data)

        t.Log(query.Sql())
        t.Log(query.Error())
    })
}
