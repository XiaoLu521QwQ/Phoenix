(fn array-foreach |array function|
    (progn
      (var i 1)
      (loop (<= i (len array))
            (progn
              (function (array-index array (- i 1)))
              (set i (+ i 1))))))

(fn array= |array1 array2|
    (progn
      (if (!= (len array1) (len array2))
          (return false))
      (var i 1)
      (loop (<= i (len array1))
            (progn
              (if (!= (array-index array1 (- i 1)) (array-index array2 (- i 1)))
                  (return false))
              (set i (+ i 1)))
            )
      true))


(fn array-insert |array item index|
    (progn
      (var i 1)
      (if (< index 0)
          (println "array-insert : cannot insert at a negative index"))
      (if (>= (+ index 1) (len array))
          (throw "array-insert : insert index is greater than the array's length")
        )
      (var returnArray [])
      (loop (<= i (len array))
            (progn
              (if (== index (- i 1))
                  (set returnArray (array-append returnArray item))
                )
              (set returnArray (array-append returnArray (array-index array (- i 1))))
              (set i (+ i 1))))
      returnArray))

; LOL
(fn array-delete |array index|
    (progn
      (var i 1)
      (if (< index 0)
          (throw "array-insert : cannot insert at a negative index")
        )
      (if (>= (+ index 1) (len array))
          (throw "array-insert : insert index is greater than the array's length")
        )
      (var returnArray [])
      (loop (<= i (len array))
            (progn
              (if (!= index (- i 1))
                  (set returnArray (array-append returnArray (array-index array (- i 1)))))
              (set i (+ i 1)))
            )
      returnArray))