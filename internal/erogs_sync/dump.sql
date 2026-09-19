-- 依 Brand ID 範圍匯出：一個 brand 對應多筆 game（不含 gameCount）
-- {{MIN_ID}} / {{MAX_ID}} 由同步功能執行時替換
WITH target_brands AS (
    SELECT
        b.id,
        b.brandname,
        b.lost
    FROM brandlist b
    WHERE b.id BETWEEN {{MIN_ID}} AND {{MAX_ID}}
) SELECT json_agg(
    json_build_object(
        'id', tb.id,
        'name', tb.brandname,
        'disband', tb.lost,
        'gamelist', COALESCE(gl.gamelist, '[]'::json)
    )
    ORDER BY tb.id
)
FROM target_brands tb
LEFT JOIN LATERAL (
    SELECT json_agg(
        json_build_object(
            'id', g.id,
            'brandErogsId', g.brandname,
            'name', g.gamename,
            'category', COALESCE(g.model, ''),
            'image', CASE
                WHEN COALESCE(g.dmm, '') <> '' THEN
                    format(
                        'https://pics.dmm.co.jp/digital/pcgame/%s/%spl.jpg',
                        g.dmm,
                        g.dmm
                    )
                ELSE ''
            END
        ) ORDER BY g.sellday DESC NULLS LAST
    ) AS gamelist
    FROM gamelist g
    WHERE g.brandname = tb.id
) gl ON TRUE;
