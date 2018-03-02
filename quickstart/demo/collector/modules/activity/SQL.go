package activity

var sqlCreate = `
insert into activity_main_info
(activity_id,
 title,
 begin_time,
 end_time,
 activity_url,
 activity_tag,
 description,
 total_amount)
values
(@activity_id,
 @title,
 to_date(@begin_time,'yyyyMMddHH24miss'),
 to_date(@end_time,'yyyyMMddHH24miss'),
 @activity_url,
 @activity_tag,
 @description,
 @total_amount )`
