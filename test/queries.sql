-- GetTodoByCategoryLabel
SELECT * from public.todo where category_id in (select id from category where label = :label);
