insert into videos (id, video_name, file_name, public, description, sort_order, thumb, converted_for_streaming, duration, is_360, created_at, updated_at)
             select id, video_name, file_name, 1, coalesce(description, ''), coalesce(sort_order, 1), thumb, 1, coalesce(duration, 0), is_360, created_at, updated_at from videos_old;


drop table videos_old;

ALTER TABLE vehicle_videos ADD FOREIGN KEY (video_id) REFERENCES videos(id);

update videos set file_name = (select REGEXP_REPLACE(file_name, '.mp4', '') );


-- finally, copy renameVideoFiles to the videos directory and run it.