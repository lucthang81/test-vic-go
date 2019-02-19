INSERT INTO admin_account (username, password, password_action)
 VALUES ('tungdt','$2a$06$TwUEzveCP1meGEsBGEyWCuI0SBWGL4t92UAZfYsyBGei1tOL0PhKy ','$2a$06$TwUEzveCP1meGEsBGEyWCuI0SBWGL4t92UAZfYsyBGei1tOL0PhKy ');
INSERT INTO admin_account (username, password, password_action)
 VALUES ('tungdt_bot','$2a$06$TwUEzveCP1meGEsBGEyWCuI0SBWGL4t92UAZfYsyBGei1tOL0PhKy ','$2a$06$TwUEzveCP1meGEsBGEyWCuI0SBWGL4t92UAZfYsyBGei1tOL0PhKy ');

INSERT INTO manager_account (username) VALUES ('daominah2');
 
INSERT INTO shop_item (name, price, discount_rate) 
  VALUES ('HTC U Ultra', 12500000, 0.25);
INSERT INTO shop_item (name, price, discount_rate) 
  VALUES ('Iphone 8', 26250000, 0.25);
INSERT INTO shop_item (name, price, discount_rate) 
  VALUES ('Iphone X', 37500000, 0.25);
INSERT INTO shop_item (name, price, discount_rate) 
  VALUES ('Samsung S9', 25000000, 0.25);
INSERT INTO shop_item (name, price, discount_rate) 
  VALUES ('Samsung S9+', 29300000, 0.25);
INSERT INTO shop_item (name, price, discount_rate) 
  VALUES ('Xe SH', 100000000, 0.25);
 
INSERT INTO vip_data (code, name,requirement_score,time_bonus_multiplier,mega_time_bonus_multiplier,leaderboard_reward_multiplier,purchase_multiplier)
 VALUES ('vip_1','ĐỒNG',0,1,1,1,1);
INSERT INTO vip_data (code, name,requirement_score,time_bonus_multiplier,mega_time_bonus_multiplier,leaderboard_reward_multiplier,purchase_multiplier)
 VALUES ('vip_2','BẠC',150,1.1,1.1,1.5,1.5);
 INSERT INTO vip_data (code, name,requirement_score,time_bonus_multiplier,mega_time_bonus_multiplier,leaderboard_reward_multiplier,purchase_multiplier)
 VALUES ('vip_3','VÀNG',4000,1.25,1.25,2,2);
 INSERT INTO vip_data (code, name,requirement_score,time_bonus_multiplier,mega_time_bonus_multiplier,leaderboard_reward_multiplier,purchase_multiplier)
 VALUES ('vip_4','BẠCH KIM',30000,1.5,1.5,2.5,2.5);
 INSERT INTO vip_data (code, name,requirement_score,time_bonus_multiplier,mega_time_bonus_multiplier,leaderboard_reward_multiplier,purchase_multiplier)
 VALUES ('vip_5','KIM CƯƠNG',250000,2,2,3,3);
 INSERT INTO vip_data (code, name,requirement_score,time_bonus_multiplier,mega_time_bonus_multiplier,leaderboard_reward_multiplier,purchase_multiplier)
 VALUES ('vip_6','KIM CƯƠNG ĐỎ',2000000,3,3,3.5,4);
 INSERT INTO vip_data (code, name,requirement_score,time_bonus_multiplier,mega_time_bonus_multiplier,leaderboard_reward_multiplier,purchase_multiplier)
 VALUES ('vip_7','KIM CƯƠNG ĐEN',8000000,4,4,4,5);

ALTER TABLE otp_code ADD COLUMN phone_number text DEFAULT '';
ALTER TABLE otp_code ADD COLUMN retry_count bigint DEFAULT 0;
ALTER TABLE otp_code ADD COLUMN data text DEFAULT '';

ALTER TABLE player ADD COLUMN is_verify boolean DEFAULT false;
ALTER TABLE player ADD COLUMN already_receive_otp_reward boolean DEFAULT false;

ALTER TABLE jackpot ADD COLUMN start_date timestamp without time zone DEFAULT (now() at time zone 'utc');
ALTER TABLE jackpot ADD COLUMN end_date timestamp without time zone DEFAULT (now() at time zone 'utc');
ALTER TABLE jackpot ADD COLUMN always_available boolean DEFAULT true;
ALTER TABLE jackpot ADD COLUMN help_text text DEFAULT '';

ALTER TABLE bacay_jackpot_record ADD COLUMN requirement bigint DEFAULT 0;

ALTER TABLE player ADD COLUMN password_change_available boolean DEFAULT false;
ALTER TABLE player ADD COLUMN phone_number_change_available boolean DEFAULT false;

ALTER TABLE jackpot ADD COLUMN start_time_daily text DEFAULT '';
ALTER TABLE jackpot ADD COLUMN end_time_daily text DEFAULT '';

ALTER TABLE admin_account ADD COLUMN admin_type character varying(25) DEFAULT 'admin';

ALTER TABLE active_record ADD COLUMN ip_address text DEFAULT '';