class Solution {
    fun groupAnagrams(strings: Array<String>): List<List<String>> {
        // create results return
        val anagrams = HashMap<STring, List<String>>()
        // for each string
        for (str in strings) {
            // sort the string alphabetically
            val sortedKey = str.toCharArray().sorted().joinToString()
            // if our HashMap contains the sort key 
            if (anagrams.containsKey(sortedKey)) {
                // append
                anagrams[sortedKey] = anagrams[sortedKey]!! + mutableListOf(str)
            } else {
                // insert
                anagrams[sortedKey] = mutableListOf(str)
            }
        }
        return anagrams.map { it.value }
    }
}