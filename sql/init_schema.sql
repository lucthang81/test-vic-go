
SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


SET search_path = public, pg_catalog;

SET default_tablespace = '';

SET default_with_oids = false;

CREATE TABLE game (
    id BIGSERIAL PRIMARY KEY,
    game_code character varying(25) NOT NULL,
    currency_type character varying(25) NOT NULL,
    data text DEFAULT '',
    help_text text DEFAULT '',
    created_at timestamp without time zone DEFAULT (now() at time zone 'utc') ,
    updated_at timestamp without time zone DEFAULT (now() at time zone 'utc') ,


    UNIQUE (game_code, currency_type)
);
ALTER TABLE public.game OWNER TO vic_user;

CREATE TABLE player (
    id BIGSERIAL PRIMARY KEY,
    username text UNIQUE,
    avatar text,
    identifier character varying(15) NOT NULL UNIQUE,
    device_identifier text DEFAULT '',
    app_type text DEFAULT '',
    phone_number text,
    password text DEFAULT '$2a$06$eZVFEhXP03wPg8J2/0z.ZuXbzNoMdUAbdN6oA5SEB.X.wzMFBWbsy',
    token character varying(15) UNIQUE,
    facebook_user_id character varying(30) UNIQUE DEFAULT NULL,

    player_type text DEFAULT 'normal',

    is_banned boolean default false,
    is_verify boolean default false,

    already_receive_otp_reward boolean default false,

    email text UNIQUE,
    password_reset_token character varying(15) DEFAULT '',
    password_change_available boolean default false,
    phone_number_change_available boolean default false,

    power bigint DEFAULT 0,
    loyalty bigint DEFAULT 0,
    level bigint DEFAULT 1,
    exp bigint DEFAULT 0,
    bet bigint DEFAULT 0,

    created_at timestamp without time zone DEFAULT (now() at time zone 'utc') ,
    updated_at timestamp without time zone DEFAULT (now() at time zone 'utc'),
    display_name text DEFAULT ''
);
ALTER TABLE public.player OWNER TO vic_user;
CREATE INDEX player_username_index ON player (username);
CREATE INDEX player_phone_number_index ON player (phone_number);
CREATE INDEX player_token_index ON player (token);
CREATE INDEX player_playerType_createdAt_index ON player (player_type, created_at);

CREATE TABLE currency (
    id BIGSERIAL PRIMARY KEY,
    player_id bigint REFERENCES player(id) ON DELETE cascade,
    currency_type character varying(25) NOT NULL,

    value bigint DEFAULT 0,

    UNIQUE (player_id, currency_type)
);
ALTER TABLE public.currency OWNER TO vic_user;


CREATE TABLE currency_type (
    id BIGSERIAL PRIMARY KEY,
    currency_type character varying(25) NOT NULL UNIQUE,
    initial_value bigint DEFAULT 0
);
ALTER TABLE public.currency_type OWNER TO vic_user;


CREATE TABLE admin_account (
    id BIGSERIAL PRIMARY KEY,
    username character varying(25) NOT NULL UNIQUE,
    password text,
    password_action text,
    token text,

    admin_type character varying(25) NOT NULL default 'admin',
    is_active boolean DEFAULT true,
    created_by_admin_id bigint REFERENCES admin_account(id),

    created_at timestamp without time zone DEFAULT (now() at time zone 'utc') ,
    updated_at timestamp without time zone DEFAULT (now() at time zone 'utc')
);
ALTER TABLE public.admin_account OWNER TO vic_user;

CREATE TABLE admin_login_activity (
    id BIGSERIAL PRIMARY KEY,
    admin_id bigint REFERENCES admin_account(id) ON DELETE cascade,
    possible_ips text,

    created_at timestamp without time zone DEFAULT (now() at time zone 'utc')
);
ALTER TABLE public.admin_login_activity OWNER TO vic_user;

CREATE TABLE achievement (
    id BIGSERIAL PRIMARY KEY,
    player_id bigint REFERENCES player(id) ON DELETE cascade,
    game_code character varying(25) NOT NULL,
    currency_type character varying(25) NOT NULL,

    win_count integer DEFAULT 0,
    lose_count integer DEFAULT 0,
    draw_count integer DEFAULT 0,
    quit_count integer DEFAULT 0,
    biggest_win bigint DEFAULT 0,

    biggest_win_this_week bigint DEFAULT 0,
    total_gain_this_week bigint DEFAULT 0,

    biggest_win_this_day bigint DEFAULT 0,
    total_gain_this_day bigint DEFAULT 0,

    created_at timestamp without time zone DEFAULT (now() at time zone 'utc') ,
    updated_at timestamp without time zone DEFAULT (now() at time zone 'utc') ,

    UNIQUE (player_id,game_code,currency_type)
);
ALTER TABLE public.achievement OWNER TO vic_user;
CREATE INDEX updated_at_index ON achievement (updated_at);

CREATE TABLE relationship (
    id BIGSERIAL PRIMARY KEY,
    from_id bigint REFERENCES player(id) ON DELETE cascade,
    to_id bigint REFERENCES player(id) ON DELETE cascade,
    relationship_type character varying(25) NOT NULL,

    created_at timestamp without time zone DEFAULT (now() at time zone 'utc') ,
    updated_at timestamp without time zone DEFAULT (now() at time zone 'utc') ,

    UNIQUE (from_id,to_id,relationship_type)
);
ALTER TABLE public.relationship OWNER TO vic_user;

CREATE TABLE friend_request (
    id BIGSERIAL PRIMARY KEY,
    from_id bigint REFERENCES player(id) ON DELETE cascade,
    to_id bigint REFERENCES player(id) ON DELETE cascade,

    created_at timestamp without time zone DEFAULT (now() at time zone 'utc') ,
    updated_at timestamp without time zone DEFAULT (now() at time zone 'utc') ,

    UNIQUE (from_id,to_id)
);
ALTER TABLE public.friend_request OWNER TO vic_user;

CREATE TABLE total_gain_weekly_prize (
    id BIGSERIAL PRIMARY KEY,
    image_url text,

    from_rank bigint DEFAULT 0,
    to_rank bigint DEFAULT 0,
    prize bigint DEFAULT 0,
    game_code character varying(25) NOT NULL
);
ALTER TABLE public.total_gain_weekly_prize OWNER TO vic_user;

CREATE TABLE gift (
    id BIGSERIAL PRIMARY KEY,
    to_id bigint REFERENCES player(id) ON DELETE cascade,
    gift_type text ,
    data text,

    status character varying(25) DEFAULT 'unread',

    currency_type character varying(25) NOT NULL,
    value bigint,

    expired_at timestamp without time zone DEFAULT (now() at time zone 'utc'),
    created_at timestamp without time zone DEFAULT (now() at time zone 'utc') ,
    updated_at timestamp without time zone DEFAULT (now() at time zone 'utc')
);
ALTER TABLE public.gift OWNER TO vic_user;

CREATE TABLE message (
    id BIGSERIAL PRIMARY KEY,
    to_id bigint REFERENCES player(id) ON DELETE cascade,
    message_type text ,
    data text,
    status character varying(25) DEFAULT 'unread',
    created_at timestamp without time zone DEFAULT (now() at time zone 'utc')
);
ALTER TABLE public.message OWNER TO vic_user;
CREATE INDEX message_i1 ON public.message
  USING btree (to_id);


CREATE TABLE time_bonus_record (
    id BIGSERIAL PRIMARY KEY,
    player_id bigint REFERENCES player(id) ON DELETE cascade,
    last_received_bonus timestamp without time zone DEFAULT (now() at time zone 'utc'),
    last_bonus_index integer DEFAULT -1,


    UNIQUE (player_id)
);
ALTER TABLE public.time_bonus_record OWNER TO vic_user;

CREATE TABLE vip_data (
    id BIGSERIAL PRIMARY KEY,
    code character varying(25) NOT NULL,
    requirement_score bigint DEFAULT 0,

    name text,
    time_bonus_multiplier real DEFAULT 1, 
    mega_time_bonus_multiplier real DEFAULT 1,
    leaderboard_reward_multiplier real DEFAULT 1,
    purchase_multiplier real DEFAULT 1
);
ALTER TABLE public.vip_data OWNER TO vic_user;

CREATE TABLE vip_record (
    id BIGSERIAL PRIMARY KEY,
    player_id bigint REFERENCES player(id) ON DELETE cascade,
    vip_code character varying(25) NOT NULL DEFAULT 'vip_1',
    vip_score bigint DEFAULT 0
);
ALTER TABLE public.vip_record OWNER TO vic_user;

CREATE TABLE app_status (
    id BIGSERIAL PRIMARY KEY,
    app_version character varying(25) NOT NULL DEFAULT '0.1',

    fake_iap boolean DEFAULT FALSE,
    fake_iap_version text DEFAULT '1.0',

    fake_iab boolean DEFAULT FALSE,
    fake_iab_version text DEFAULT '1.0',
    
    maintenance_status boolean default FALSE,
    maintenance_start timestamp without time zone DEFAULT NULL ,
    maintenance_end timestamp without time zone DEFAULT NULL
);
ALTER TABLE public.app_status OWNER TO vic_user;

CREATE TABLE payment_requirement (
    id BIGSERIAL PRIMARY KEY,
    min_money_left bigint DEFAULT 20000,
    min_days_since_purchase bigint DEFAULT 14,
    min_total_bet bigint DEFAULT 50000,
    purchase_multiplier bigint DEFAULT 11,
    max_payment_count_day bigint DEFAULT 3,
    rule_text text DEFAULT ''
);
ALTER TABLE public.payment_requirement OWNER TO vic_user;

CREATE TABLE feedback (
    id BIGSERIAL PRIMARY KEY,
    player_id bigint REFERENCES player(id) ON DELETE cascade,
    star integer DEFAULT 0,
    feedback text DEFAULT '',
    version character varying(25) NOT NULL DEFAULT '0.1',
    created_at timestamp without time zone DEFAULT (now() at time zone 'utc') ,
    updated_at timestamp without time zone DEFAULT (now() at time zone 'utc')
);
ALTER TABLE public.feedback OWNER TO vic_user;


CREATE TABLE event (
    id BIGSERIAL PRIMARY KEY,
    icon_url text DEFAULT '',
    event_type character varying(25) NOT NULL DEFAULT 'one_time',
    title text DEFAULT '',
    description text DEFAULT '',
    tip_title text DEFAULT '',
    tip_description text DEFAULT '',
    priority integer DEFAULT 0,
    data text DEFAULT ''
);
ALTER TABLE public.event OWNER TO vic_user;

/*
record
*/
CREATE TABLE active_record (
    id BIGSERIAL PRIMARY KEY,
    player_id bigint REFERENCES player(id) ON DELETE cascade,
    device_code text DEFAULT '',
    device_type text DEFAULT '',
    ip_address text DEFAULT '',

    start_date timestamp without time zone DEFAULT (now() at time zone 'utc'),
    end_date timestamp without time zone DEFAULT (now() at time zone 'utc')
);
ALTER TABLE public.active_record OWNER TO vic_user;
CREATE INDEX active_record_end_date_index ON active_record (start_date);
CREATE INDEX active_record_start_date_index ON active_record (end_date);
CREATE INDEX active_record_ip_address_index ON active_record (ip_address);

CREATE TABLE match_record (
    id BIGSERIAL PRIMARY KEY,
    tax bigint DEFAULT 0,
    bet bigint DEFAULT 0,

    win bigint DEFAULT 0,
    lose bigint DEFAULT 0,

    bot_win bigint DEFAULT 0,
    bot_lose bigint DEFAULT 0,

    match_data text DEFAULT '',

    game_code character varying(25) NOT NULL,
    currency_type character varying(25) NOT NULL,
    requirement bigint DEFAULT 0,

    minah_id character varying(100),
    
    created_at timestamp without time zone DEFAULT (now() at time zone 'utc')
);
ALTER TABLE public.match_record OWNER TO vic_user;
CREATE INDEX match_record_game_code_index ON match_record (game_code);
CREATE INDEX match_record_created_at_index ON match_record (created_at);
CREATE INDEX match_record_currency_type_index ON match_record (currency_type);
CREATE INDEX match_record_minah_id_index ON match_record (minah_id);

CREATE TABLE player_match_record (
    id BIGSERIAL PRIMARY KEY,
    player_id bigint REFERENCES player(id) ON DELETE cascade,
    match_record_id bigint REFERENCES match_record(id) ON DELETE cascade,

    UNIQUE (player_id, match_record_id)
);
ALTER TABLE public.player_match_record OWNER TO vic_user;

CREATE TABLE purchase_record (
    id BIGSERIAL PRIMARY KEY,
    player_id bigint REFERENCES player(id) ON DELETE cascade,
    purchase bigint DEFAULT 0,
    purchase_type character varying(25) DEFAULT 'paybnb',
    card_code text,
    currency_type character varying(25),
    transaction_id text DEFAULT '',
    value_before bigint DEFAULT 0,
    value_after bigint DEFAULT 0,

    created_at timestamp without time zone DEFAULT (now() at time zone 'utc')
);
ALTER TABLE public.purchase_record OWNER TO vic_user;
CREATE INDEX purchase_record_player_id_index ON purchase_record (player_id);
CREATE INDEX purchase_record_created_at_index ON purchase_record (created_at);
CREATE INDEX purchase_record_purchase_type_index ON purchase_record (purchase_type);
CREATE INDEX purchase_record_tran_index ON purchase_record (transaction_id);
ALTER TABLE  public.purchase_record ADD real_money_value bigint DEFAULT 0;

CREATE TABLE purchase_referer (
    id BIGSERIAL PRIMARY KEY,
    transaction_id text DEFAULT '',
    player_id bigint REFERENCES player(id),
    purchase_type character varying(25) DEFAULT 'paybnb',
    card_code text,
    card_serial text,

    created_at timestamp without time zone DEFAULT (now() at time zone 'utc')
);
ALTER TABLE public.purchase_referer OWNER TO vic_user;
CREATE INDEX purchase_referer_transaction_index ON purchase_referer (transaction_id);
CREATE INDEX purchase_referer_created_at_index ON purchase_referer (created_at);

CREATE TABLE card (
    id BIGSERIAL PRIMARY KEY,
    card_type text DEFAULT '',
    card_code text DEFAULT '',
    serial_code text DEFAULT '' UNIQUE,
    card_number text DEFAULT '' UNIQUE,
    card_value text DEFAULT '',
    status character varying(25) NOT NULL DEFAULT 'unclaimed',

    claimed_by_player_id bigint REFERENCES player(id),
    accepted_by_admin_id bigint REFERENCES admin_account(id),


    created_at timestamp without time zone DEFAULT (now() at time zone 'utc'),
    claimed_at timestamp without time zone DEFAULT (now() at time zone 'utc')
);
ALTER TABLE public.card OWNER TO vic_user;
CREATE UNIQUE INDEX card_data_unique ON card (card_number, serial_code);

CREATE TABLE bank (
    id BIGSERIAL PRIMARY KEY,
    value bigint DEFAULT 0,
    currency_type character varying(25),
    game_code character varying(25) NOT NULL,

    UNIQUE(currency_type, game_code)
);
ALTER TABLE public.bank OWNER TO vic_user;
CREATE UNIQUE INDEX currency_type_game_code_bank ON bank(currency_type, game_code);

CREATE TABLE bank_record (
    id BIGSERIAL PRIMARY KEY,
    match_id bigint REFERENCES match_record(id),
    player_id bigint REFERENCES player(id),
    game_code character varying(25) NOT NULL,
    currency_type character varying(25),

    value_before bigint DEFAULT 0,
    value_after bigint DEFAULT 0,

    created_at timestamp without time zone DEFAULT (now() at time zone 'utc')
);
ALTER TABLE public.bank_record OWNER TO vic_user;
CREATE INDEX bank_record_game_code_index ON bank_record (game_code);
CREATE INDEX bank_record_created_at_index ON bank_record (created_at);

CREATE TABLE payment_record (
    id BIGSERIAL PRIMARY KEY,
    player_id bigint REFERENCES player(id) ON DELETE cascade,
    payment bigint DEFAULT 0,
    currency_type character varying(25) NOT NULL,
    value_before bigint DEFAULT 0,
    value_after bigint DEFAULT 0,
    tax bigint DEFAULT 0,
    payment_type character varying(25) DEFAULT 'card',

    status character varying(25) NOT NULL DEFAULT '',
    replied_by_admin_id bigint REFERENCES admin_account(id),

    data text DEFAULT '',

    card_code text DEFAULT '',
    card_id bigint REFERENCES card(id),

    created_at timestamp without time zone DEFAULT (now() at time zone 'utc'),
    replied_at timestamp without time zone DEFAULT (now() at time zone 'utc')
);
ALTER TABLE public.payment_record OWNER TO vic_user;
CREATE INDEX payment_record_player_id_index ON payment_record (player_id);
CREATE INDEX payment_record_created_at_index ON payment_record (created_at);
CREATE INDEX payment_record_currency_type_index ON payment_record (currency_type);


CREATE TABLE currency_record (
    id BIGSERIAL PRIMARY KEY,
    player_id bigint REFERENCES player(id) ON DELETE cascade,
    action text DEFAULT '',
    game_code text DEFAULT '',
    change bigint DEFAULT 0,
    currency_type character varying(25) NOT NULL,
    value_before bigint DEFAULT 0,
    value_after bigint DEFAULT 0,
    additional_data text default '',

    created_at timestamp without time zone DEFAULT (now() at time zone 'utc')
);
ALTER TABLE public.currency_record OWNER TO vic_user;
CREATE INDEX currency_record_pId_moneyType_time_index ON currency_record USING  btree (player_id, currency_type, created_at);



CREATE TABLE ccu_record (
    id BIGSERIAL PRIMARY KEY,
    online_total_count bigint DEFAULT 0,
    online_bot_count bigint DEFAULT 0,
    online_normal_count bigint DEFAULT 0,

    game_online_data text default '',
    created_at timestamp without time zone DEFAULT (now() at time zone 'utc')
);
ALTER TABLE public.ccu_record OWNER TO vic_user;
CREATE INDEX cc_record_created_at_index ON ccu_record (created_at);


CREATE TABLE card_type (
    id BIGSERIAL PRIMARY KEY,
    card_code text DEFAULT '' UNIQUE,
    card_name text DEFAULT '',
    money bigint DEFAULT 0,

    created_at timestamp without time zone DEFAULT (now() at time zone 'utc')
);
ALTER TABLE public.card_type OWNER TO vic_user;

CREATE TABLE purchase_type (
    id BIGSERIAL PRIMARY KEY,
    purchase_code text DEFAULT '',
    purchase_type text DEFAULT '',
    money bigint DEFAULT 0,

    created_at timestamp without time zone DEFAULT (now() at time zone 'utc')
);
ALTER TABLE public.purchase_type OWNER TO vic_user;

CREATE TABLE pn_data (
    id BIGSERIAL PRIMARY KEY,
    apns_keyfile_content text DEFAULT '',
    apns_cerfile_content text DEFAULT '',
    apns_type character varying(25) DEFAULT '',
    gcm_api_key text DEFAULT '',
    app_type text DEFAULT 'bighero' UNIQUE
);
ALTER TABLE public.pn_data OWNER TO vic_user;

CREATE TABLE pn_device (
    id BIGSERIAL PRIMARY KEY,
    player_id bigint REFERENCES player(id) ON DELETE cascade UNIQUE,
    apns_device_token text DEFAULT '',
    gcm_device_token text DEFAULT '',

    created_at timestamp without time zone DEFAULT (now() at time zone 'utc')
);
ALTER TABLE public.pn_device OWNER TO vic_user;

CREATE TABLE pn_schedule (
    id BIGSERIAL PRIMARY KEY,
    time timestamp without time zone DEFAULT (now() at time zone 'utc'),
    message text DEFAULT ''
);
ALTER TABLE public.pn_schedule OWNER TO vic_user;

CREATE TABLE popup_message (
    id BIGSERIAL PRIMARY KEY,
    title text DEFAULT '',
    content text DEFAULT ''
);
ALTER TABLE public.popup_message OWNER TO vic_user;


/* jackpot */
CREATE TABLE jackpot (
    id BIGSERIAL PRIMARY KEY,
    code text DEFAULT '',

    currency_type character varying(25) NOT NULL,
    value bigint DEFAULT 0,
    
    start_date timestamp without time zone DEFAULT (now() at time zone 'utc'),
    end_date timestamp without time zone DEFAULT (now() at time zone 'utc'),
    always_available boolean DEFAULT true,
    help_text text DEFAULT '',
    start_time_daily text DEFAULT '',
    end_time_daily text DEFAULT '',

    UNIQUE(code, currency_type)
);
ALTER TABLE public.jackpot OWNER TO vic_user;
CREATE INDEX jackpot_currency_type_index ON jackpot (currency_type);
CREATE INDEX jackpot_code_index ON jackpot(code);

CREATE TABLE public.jackpot_hit_record
(
  id bigserial,
  gamecode character varying(30),
  money_amount bigint,
  hit_player_id bigint,
  hit_player_username character varying(100),
  created_at timestamp without time zone DEFAULT timezone('utc'::text, now()),
  CONSTRAINT jackpot_hit_pkey PRIMARY KEY (id)
)
WITH (
  OIDS=FALSE
);
ALTER TABLE public.jackpot_hit_record
  OWNER TO vic_user;
CREATE INDEX jackpot_hit_record_gamecode_createdAt ON jackpot_hit_record USING  btree (gamecode, created_at);

  
/*
 * vip point 
 * */
CREATE TABLE vip_point_data (
    id BIGSERIAL PRIMARY KEY,
    vip_point_rate bigint DEFAULT 2000
);
ALTER TABLE public.vip_point_data OWNER TO vic_user;

CREATE TABLE gift_payment_type (
    id BIGSERIAL PRIMARY KEY,
    code text DEFAULT '',
    name text DEFAULT '',
    value bigint DEFAULT 0,
    quantity bigint DEFAULT 0,
    image_url text DEFAULT '',

    created_at timestamp without time zone DEFAULT (now() at time zone 'utc')
);
ALTER TABLE public.gift_payment_type OWNER TO vic_user;

/* tracking */
CREATE TABLE maubinh_type_record (
    id BIGSERIAL PRIMARY KEY,
    match_id bigint REFERENCES match_record(id) ON DELETE cascade,
    player_id bigint REFERENCES player(id) ON DELETE cascade,
    currency_type character varying(25) NOT NULL,
    player_type text DEFAULT 'normal',
    type text DEFAULT '',
    position text DEFAULT '',
    cards text DEFAULT '',

    created_at timestamp without time zone DEFAULT (now() at time zone 'utc')
);
ALTER TABLE public.maubinh_type_record OWNER TO vic_user;
CREATE INDEX maubinh_type_record_currency_type_index ON maubinh_type_record (currency_type);
CREATE INDEX maubinh_type_record_player_type_index ON maubinh_type_record (player_type);
CREATE INDEX maubinh_type_record_type_index ON maubinh_type_record (type);
CREATE INDEX maubinh_type_record_position_index ON maubinh_type_record (position);
CREATE INDEX maubinh_type_record_created_at_index ON maubinh_type_record (created_at);

CREATE TABLE maubinh_white_win_record (
    id BIGSERIAL PRIMARY KEY,
    match_id bigint REFERENCES match_record(id) ON DELETE cascade,
    player_id bigint REFERENCES player(id) ON DELETE cascade,
    currency_type character varying(25) NOT NULL,
    player_type text DEFAULT 'normal',
    white_win_type text DEFAULT '',
    cards text DEFAULT '',

    created_at timestamp without time zone DEFAULT (now() at time zone 'utc')
);
ALTER TABLE public.maubinh_white_win_record OWNER TO vic_user;
CREATE INDEX maubinh_white_win_record_currency_type_index ON maubinh_white_win_record (currency_type);
CREATE INDEX maubinh_white_win_record_player_type_index ON maubinh_white_win_record (player_type);
CREATE INDEX maubinh_white_win_record_white_win_type_index ON maubinh_white_win_record (white_win_type);
CREATE INDEX maubinh_white_win_record_created_at_index ON maubinh_white_win_record (created_at);


CREATE TABLE phone_auth_server_number (
    phone_number character varying(25) PRIMARY KEY
);

CREATE TABLE public.otp_code
(
  id BIGSERIAL PRIMARY KEY,
  player_id bigint,
  phone_number text DEFAULT ''::text,
  otp_code text DEFAULT ''::text,
  reason text DEFAULT ''::text,
  data text DEFAULT ''::text,
  status character varying(25) NOT NULL,
  retry_count bigint DEFAULT 0,
  expired_at timestamp without time zone DEFAULT timezone('utc'::text, now()),
  created_at timestamp without time zone DEFAULT timezone('utc'::text, now()),
  passwd text DEFAULT 123456,
  CONSTRAINT otp_code_player_id_fkey FOREIGN KEY (player_id)
      REFERENCES public.player (id) MATCH SIMPLE
      ON UPDATE NO ACTION ON DELETE CASCADE
)
WITH (
  OIDS=FALSE
);
ALTER TABLE public.otp_code
  OWNER TO vic_user;
CREATE INDEX otp_code_created_at_index
  ON public.otp_code
  USING btree
  (created_at);
CREATE INDEX otp_code_reason_index
  ON public.otp_code
  USING btree
  (reason COLLATE pg_catalog."default");

  
CREATE TABLE public.match_statistic
(
  id BIGSERIAL PRIMARY KEY,
  player_id bigint,
  win bigint DEFAULT 0,
  lose bigint DEFAULT 0,
  draw bigint DEFAULT 0,
  "time" bigint DEFAULT 0,
  won_money bigint DEFAULT 0,
  lost_money bigint DEFAULT 0,
  game_code text,
  currency_type text,
  username text,
  time_type text
)
WITH (
  OIDS=FALSE
);
ALTER TABLE public.match_statistic
  OWNER TO vic_user;

  
CREATE TABLE public.ingame_global_text
(
  id BIGSERIAL,
  data text DEFAULT '{}'::text,
  created_at timestamp without time zone DEFAULT timezone('utc'::text, now()),
  priority bigint DEFAULT 1,
  CONSTRAINT ingame_global_text_pkey PRIMARY KEY (id)
)
WITH (
  OIDS=FALSE
);
ALTER TABLE public.ingame_global_text
  OWNER TO vic_user;
CREATE INDEX ingame_global_text_priority_created_at_idx
  ON public.ingame_global_text
  USING btree
  (priority, created_at);


CREATE TABLE public.player_source
(
  "player_id" bigint PRIMARY KEY,
  "register_platform" character varying(100),
  "register_partner" character varying(100),
  "register_time" timestamp without time zone,
  "last_login_platform" character varying(100),
  "last_login_partner" character varying(100),
  "last_login_time" timestamp without time zone,
  "is_iap_on" boolean NOT NULL DEFAULT true,
  "is_card_pay_on" boolean NOT NULL DEFAULT false,
  "is_store_tester" boolean NOT NULL DEFAULT false
)
WITH (
  OIDS=FALSE
);
ALTER TABLE public.player_source
  OWNER TO vic_user;

CREATE TABLE currency_sum (
    id BIGSERIAL PRIMARY KEY,
    sum_users_money bigint,
    created_at timestamp without time zone DEFAULT (now() at time zone 'utc')
);
ALTER TABLE public.currency_sum OWNER TO vic_user;


CREATE TABLE public.purchase_first_time
(
  player_id bigint NOT NULL,
  datetime timestamp without time zone,
  CONSTRAINT purchase_first_time_pkey PRIMARY KEY (player_id)
)
WITH (
  OIDS=FALSE
);
ALTER TABLE public.purchase_first_time
  OWNER TO vic_user;
CREATE INDEX purchase_first_time_datetime_idx
  ON public.purchase_first_time
  USING btree
  (datetime);
  
CREATE TABLE public.purchase_first_time_daily
(
  player_id bigint NOT NULL,
  date_s text DEFAULT ''::text,
  CONSTRAINT purchase_first_time_daily_pkey PRIMARY KEY (player_id, date_s)
)
WITH (
  OIDS=FALSE
);
ALTER TABLE public.purchase_first_time_daily
  OWNER TO vic_user;
  


CREATE TABLE public.cash_out_record
(
  id BIGSERIAL,
  player_id bigint,
  action text DEFAULT ''::text,
  game_code text DEFAULT ''::text,
  change bigint DEFAULT 0,
  currency_type character varying(25) NOT NULL,
  value_before bigint DEFAULT 0,
  value_after bigint DEFAULT 0,
  additional_data text DEFAULT ''::text,
  created_at timestamp without time zone DEFAULT timezone('utc'::text, now()),
  is_verified_by_admin boolean DEFAULT false,
  is_paid boolean DEFAULT false,
  transaction_id text,
  verified_time timestamp without time zone,
  CONSTRAINT cash_out_record_pkey PRIMARY KEY (id),
  CONSTRAINT cash_out_record_player_id_fkey FOREIGN KEY (player_id)
      REFERENCES public.player (id) MATCH SIMPLE
      ON UPDATE NO ACTION ON DELETE CASCADE
)
WITH (
  OIDS=FALSE
);
ALTER TABLE public.cash_out_record
  OWNER TO vic_user;
CREATE INDEX cash_out_record_is_verified_by_admin_created_at_idx
  ON public.cash_out_record
  USING btree
  (is_verified_by_admin, created_at);
CREATE INDEX cash_out_record_pid_moneytype_time_index
  ON public.cash_out_record
  USING btree
  (player_id, currency_type COLLATE pg_catalog."default", created_at);
ALTER TABLE  public.cash_out_record ADD real_money_value bigint DEFAULT 0;

  

CREATE TABLE public.player_logins_track
(
  player_id bigint NOT NULL,
  n_continuous_logins bigint NOT NULL DEFAULT 1,
  is_logged_in_today boolean NOT NULL DEFAULT true,
  last_gift3_time timestamp without time zone NOT NULL DEFAULT '2000-01-01 00:00:00'::timestamp without time zone,
  last_gift7_time timestamp without time zone NOT NULL DEFAULT '2000-01-01 00:00:00'::timestamp without time zone,
  CONSTRAINT player_logins_track_pkey PRIMARY KEY (player_id)
)
WITH (
  OIDS=FALSE
);
ALTER TABLE public.player_logins_track
  OWNER TO vic_user;

  
CREATE TABLE public.event_top_result
(
  id BIGSERIAL NOT NULL,
  event_name text,
  starting_time timestamp without time zone,
  finishing_time timestamp without time zone,
  map_position_to_prize text,
  full_order text,
  is_paid boolean,
  CONSTRAINT event_top_result_pkey PRIMARY KEY (id)
)
WITH (
  OIDS=FALSE
);
ALTER TABLE public.event_top_result
  OWNER TO vic_user;  
CREATE INDEX event_top_result_is_paid_index
  ON public.event_top_result
  USING btree
  (is_paid);
  
  
CREATE TABLE public.player_lucky_number
(
  id BIGSERIAL NOT NULL,
  player_id bigint,
  number bigint,
  valid_date text,
  prize bigint,
  is_hit boolean NOT NULL DEFAULT false,
  CONSTRAINT player_lucky_number_pkey PRIMARY KEY (id)
)
WITH (
  OIDS=FALSE
);
ALTER TABLE public.player_lucky_number
  OWNER TO vic_user;
CREATE INDEX player_lucky_number_index1
  ON public.player_lucky_number
  USING btree
  (valid_date COLLATE pg_catalog."default");
  
  
CREATE TABLE public.player_privileges
(
  "player_id" bigint PRIMARY KEY,
  "can_create_room" boolean NOT NULL DEFAULT false,
  "can_transfer_money" boolean NOT NULL DEFAULT false,
  "can_receive_money" boolean NOT NULL DEFAULT false
)
WITH (
  OIDS=FALSE
);
ALTER TABLE public.player_privileges
  OWNER TO vic_user;
  
CREATE TABLE public.player_agency
(
  "player_id" bigint PRIMARY KEY,
  "is_accepted" boolean DEFAULT false,
  "bank_name" text DEFAULT ''::text,
  "bank_account_number" text DEFAULT ''::text,
  "bank_account__name" text DEFAULT ''::text,
  "address"  text DEFAULT ''::text,
  "rate_kim_to_vnd" double precision DEFAULT 1.3,
  "email" text DEFAULT ''::text,
  "skype" text DEFAULT ''::text,
  "is_hidding" boolean DEFAULT false
)
WITH (
  OIDS=FALSE
);
ALTER TABLE public.player_agency
  OWNER TO vic_user;
ALTER TABLE public.player_agency ADD "phone2" text DEFAULT ''::text; 
  
  
CREATE TABLE public.player_transfer_record
(
    id BIGSERIAL PRIMARY KEY,
    sender_id bigint DEFAULT 0,
    target_id  bigint DEFAULT 0,
    amount_kim bigint DEFAULT 0,
    created_time timestamp without time zone DEFAULT (now() at time zone 'utc'),
    has_sender_checked boolean DEFAULT false,
    has_target_checked boolean DEFAULT false
)
WITH (
  OIDS=FALSE
);
ALTER TABLE public.player_transfer_record
  OWNER TO vic_user;
CREATE INDEX player_transfer_record_i1 ON public.player_transfer_record
  USING btree (sender_id, created_time);
CREATE INDEX player_transfer_record_i2 ON public.player_transfer_record
  USING btree (target_id, created_time);
  
CREATE TABLE message_with_admin (
    id BIGSERIAL PRIMARY KEY,
    player_id bigint,
    is_from_user boolean,
    has_read boolean DEFAULT false,
    message text,
    created_at timestamp without time zone DEFAULT (now() at time zone 'utc')
);
ALTER TABLE public.message_with_admin OWNER TO vic_user;
CREATE INDEX message_with_admin_pid_time_index_index ON public.message_with_admin
  USING btree (player_id, created_at);
CREATE INDEX message_with_admin_has_read_player_id_index ON public.message_with_admin
  USING btree (has_read, player_id);

CREATE TABLE message_with_admin_counter (
    player_id bigint PRIMARY KEY,
    n_remaining_message bigint,
    last_time timestamp without time zone DEFAULT (now() at time zone 'utc')
);
ALTER TABLE public.message_with_admin_counter OWNER TO vic_user;

-- Table: public.gift_code
-- DROP TABLE public.gift_code;
CREATE TABLE public.gift_code
(
  id BIGSERIAL PRIMARY KEY,
  code text DEFAULT ''::text,
  name text DEFAULT ''::text,
  value bigint DEFAULT 0,
  quantity bigint DEFAULT 0,
  image_url text DEFAULT ''::text,
  created_at timestamp without time zone DEFAULT timezone('utc'::text, now()),
  player_id bigint DEFAULT 0, -- nguoi su dung giftcode
  status integer DEFAULT 0, -- = 0 la chua dung...
  expire_at timestamp without time zone DEFAULT timezone('utc'::text, now()), -- thoi gian het han su dung
  current_type text, -- money ...
  not_reuse bigint, -- not_reuse =0 cho phep 1 username nhap nhieu gift code...
  CONSTRAINT gift_code_code_key UNIQUE (code)
)
WITH (
  OIDS=FALSE
);
ALTER TABLE public.gift_code
  OWNER TO vic_user;

CREATE TABLE public.gift_code_log
(
  code text DEFAULT ''::text,
  player_id bigint DEFAULT 0,
  used_at timestamp without time zone,
  id BIGSERIAL PRIMARY KEY
)
WITH (
  OIDS=FALSE
);
ALTER TABLE public.gift_code_log
  OWNER TO vic_user;
-- Index: public.scrd
-- DROP INDEX public.scrd;
CREATE UNIQUE INDEX scrd
  ON public.gift_code_log
  USING btree
  (code COLLATE pg_catalog."default", player_id NULLS FIRST);

CREATE TABLE public.gift_code_percentage
(
  id BIGSERIAL PRIMARY KEY,
  code text DEFAULT ''::text,
  player_id bigint DEFAULT 0,
  percentage real DEFAULT 0,
  has_inputted_code boolean DEFAULT false,
  has_charged_money boolean DEFAULT false,
  unique_key bigint DEFAULT 0,
  created_time timestamp without time zone DEFAULT timezone('utc'::text, now()),
  inputted_time timestamp without time zone DEFAULT timezone('utc'::text, now()),
  charged_time timestamp without time zone DEFAULT timezone('utc'::text, now())
)
WITH (
  OIDS=FALSE
);
ALTER TABLE public.gift_code_percentage
  OWNER TO vic_user;
CREATE INDEX gift_code_percentage_i1 ON public.gift_code_percentage
  USING btree (code);
CREATE INDEX gift_code_percentage_i2 ON public.gift_code_percentage
  USING btree (player_id);
  
CREATE TABLE public.event_collecting_pieces_result
(
  id BIGSERIAL NOT NULL,
  event_name text,
  starting_time timestamp without time zone,
  finishing_time timestamp without time zone,
  n_pieces_to_complete int,
  n_limit_prizes int,
  n_rare_pieces int,
  map_pid_to_map_pieces text,
  is_paid boolean,
  CONSTRAINT eventcp_result_pkey PRIMARY KEY (id)
)
WITH (
  OIDS=FALSE
);
ALTER TABLE public.event_collecting_pieces_result
  OWNER TO vic_user;  
CREATE INDEX eventcp_result_is_paid_index
  ON public.event_collecting_pieces_result
  USING btree
  (is_paid);

  
CREATE TABLE public.gift_code_random
(
  id BIGSERIAL PRIMARY KEY,
  code text DEFAULT ''::text,
  player_id bigint DEFAULT 0,
  map_money text DEFAULT ''::text,
  has_inputted_code boolean DEFAULT false,
  money_result bigint DEFAULT 0,
  unique_key bigint DEFAULT 0,
  created_time timestamp without time zone DEFAULT timezone('utc'::text, now()),
  inputted_time timestamp without time zone DEFAULT timezone('utc'::text, now())
)
WITH (
  OIDS=FALSE
);
ALTER TABLE public.gift_code_random
  OWNER TO vic_user;
CREATE INDEX gift_code_random_i1 ON public.gift_code_random
  USING btree (code);
CREATE INDEX gift_code_random_i2 ON public.gift_code_random
  USING btree (player_id);

  
CREATE TABLE public.shop_item
(
  id BIGSERIAL PRIMARY KEY,
  name text DEFAULT ''::text,
  price bigint DEFAULT 0,
  discount_rate real DEFAULT 0,
  url text DEFAULT ''::text,
  addditional_data text DEFAULT ''::text,
  created_time timestamp without time zone DEFAULT timezone('utc'::text, now())
)
WITH (
  OIDS=FALSE
);
ALTER TABLE public.shop_item
  OWNER TO vic_user;
  
CREATE TABLE public.shop_item_buyer
(
  id BIGSERIAL PRIMARY KEY,
  item_id bigint,
  buyer_id bigint,
  buyer_name text DEFAULT ''::text,
  buyer_phone text DEFAULT ''::text,
  buyer_address text DEFAULT ''::text,
  addditional_data text DEFAULT ''::text,
  created_time timestamp without time zone DEFAULT timezone('utc'::text, now())
)
WITH (
  OIDS=FALSE
);
ALTER TABLE public.shop_item_buyer
  OWNER TO vic_user;
  
  
CREATE TABLE public.purchase_record_bank
(
  id BIGSERIAL PRIMARY KEY,
  player_id bigint DEFAULT 0,
  amount_vnd double precision DEFAULT 0,
  amount_myr double precision DEFAULT 0,
  amount_kim double precision DEFAULT 0,
  kim_before double precision DEFAULT 0,
  kim_after double precision DEFAULT 0,
  paytrust88_data text DEFAULT ''::text,
  created_time timestamp without time zone DEFAULT timezone('utc'::text, now())
)
WITH (
  OIDS=FALSE
);
ALTER TABLE public.purchase_record_bank
  OWNER TO vic_user;
CREATE INDEX purchase_record_bank_i1 ON public.purchase_record_bank
  USING btree (player_id, created_time);
  
  
CREATE TABLE public.cash_out_paytrust88_record
(
  id BIGSERIAL PRIMARY KEY,
  player_id bigint DEFAULT 0,
  bank_name text DEFAULT ''::text,
  bank_account_number text DEFAULT ''::text,
  amount_vnd double precision DEFAULT 0,
  amount_kim double precision DEFAULT 0,
  kim_before double precision DEFAULT 0,
  kim_after double precision DEFAULT 0,
  paytrust88_data text DEFAULT ''::text,
  created_at timestamp without time zone DEFAULT timezone('utc'::text, now()),
  is_verified_by_admin boolean DEFAULT false,
  is_paid boolean DEFAULT false,
  verified_time timestamp without time zone DEFAULT timezone('utc'::text, now())
)
WITH (
  OIDS=FALSE
);
ALTER TABLE public.cash_out_paytrust88_record
  OWNER TO vic_user;
CREATE INDEX cash_out_paytrust88_record_is_verified_by_admin_created_at_idx
  ON public.cash_out_paytrust88_record
  USING btree
  (is_verified_by_admin, created_at);
CREATE INDEX cash_out_paytrust88_record_pid_time_index
  ON public.cash_out_paytrust88_record
  USING btree
  (player_id, created_at);
  
 
CREATE TABLE public.zkey_value
(
  zkey text PRIMARY KEY,
  zvalue text DEFAULT ''::text,
  last_modified timestamp without time zone DEFAULT timezone('utc'::text, now())
)
WITH (
  OIDS=FALSE
);
ALTER TABLE public.zkey_value
  OWNER TO vic_user;
  
CREATE TABLE public.zglobal_var
(
  zkey text PRIMARY KEY,
  zvalue text DEFAULT ''::text,
  last_modified timestamp without time zone DEFAULT timezone('utc'::text, now())
)
WITH (
  OIDS=FALSE
);
ALTER TABLE public.zglobal_var 
  OWNER TO vic_user;
  
  
CREATE TABLE public.purchase_record_iap_android
(
  order_id text PRIMARY KEY,
  player_id bigint DEFAULT 0,
  receipt text DEFAULT ''::text,
  created_time timestamp without time zone DEFAULT timezone('utc'::text, now())
)
WITH (
  OIDS=FALSE
);
ALTER TABLE public.purchase_record_iap_android
  OWNER TO vic_user;
CREATE INDEX purchase_record_iap_android_i5 ON public.purchase_record_iap_android
  USING btree (player_id, created_time);
CREATE INDEX purchase_record_iap_android_i6 ON public.purchase_record_iap_android
  USING btree (created_time);
  
  
CREATE TABLE public.manager_account
(
  id bigserial,
  username text,
  password text DEFAULT '123qwe'::text,
  login_token text,
  created_at timestamp without time zone DEFAULT timezone('utc'::text, now()),
  partner text NOT NULL DEFAULT 'all'::text,
  CONSTRAINT manager_account_pkey PRIMARY KEY (id)
)
WITH (
  OIDS=FALSE
);
ALTER TABLE public.manager_account
  OWNER TO vic_user;
CREATE INDEX index_username
  ON public.manager_account
  USING btree
  (username COLLATE pg_catalog."default");


CREATE TABLE public.purchase_sum
(
	player_id BIGINT DEFAULT 0,
	purchase_type TEXT DEFAULT '',
	sum_value BIGINT DEFAULT 0,
    CONSTRAINT purchase_sum_pkey PRIMARY KEY (player_id, purchase_type)
);
ALTER TABLE public.purchase_sum OWNER TO vic_user;


--
CREATE TABLE public.rank (
    rank_id BIGSERIAL, CONSTRAINT rank_pkey PRIMARY KEY (rank_id),
    rank_name TEXT DEFAULT '' UNIQUE,
    started_time TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
INSERT INTO public.rank (rank_name) VALUES ('Test1');
INSERT INTO public.rank (rank_name) VALUES ('Test2');
INSERT INTO public.rank (rank_name) VALUES ('Net worth');
INSERT INTO public.rank (rank_name) VALUES ('Number of wins');




--
CREATE TABLE public.rank_key (
    rank_id BIGINT DEFAULT 0 REFERENCES public.rank (rank_id),
    user_id BIGINT DEFAULT 0,
    CONSTRAINT rank_key_pkey PRIMARY KEY (rank_id, user_id),
    rkey DOUBLE PRECISION DEFAULT 0,
    last_modified TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX rank_key_i01 ON public.rank_key USING btree
    (rank_id, rkey, last_modified, user_id);


--
CREATE TABLE public.rank_archive (
    archive_id BIGSERIAL, CONSTRAINT rank_archive_pkey PRIMARY KEY (archive_id),
    rank_id BIGINT DEFAULT 0,
    rank_name TEXT DEFAULT '',
    started_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    finished_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    top_10 TEXT DEFAULT '[]',
    full_order TEXT DEFAULT '[]'
);


--
CREATE TABLE top_taixiu (
    top_date text DEFAULT '',
    player_id BIGINT DEFAULT 0,
    CONSTRAINT top_taixiu_pkey PRIMARY KEY (top_date, player_id),
    current_win_streak BIGINT DEFAULT 0,
    peak_win_streak BIGINT DEFAULT 0,
    current_loss_streak BIGINT DEFAULT 0,
    peak_loss_streak BIGINT DEFAULT 0,
    change_money BIGINT DEFAULT 0,
    start_money BIGINT DEFAULT 0,
    finish_money BIGINT DEFAULT 0
);
CREATE INDEX top_taixiu_i01 ON top_taixiu USING btree
    (top_date, change_money);