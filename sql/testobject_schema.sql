--
-- PostgreSQL database dump
--

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET search_path = public, pg_catalog;

SET default_tablespace = '';

SET default_with_oids = false;


CREATE TABLE avahardcore_testobject (
    id SERIAL PRIMARY KEY,
    username character varying(25) NOT NULL,
    avatar text,
    identifier character varying(15) NOT NULL,
    token character varying(15),

    power bigint DEFAULT 0,
    loyalty bigint DEFAULT 0,

    created_at timestamp without time zone DEFAULT (now() at time zone 'utc') ,
    updated_at timestamp without time zone DEFAULT (now() at time zone 'utc')
);

ALTER TABLE public.avahardcore_testobject OWNER TO vic_user;

